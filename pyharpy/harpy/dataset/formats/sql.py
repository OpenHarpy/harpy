from dataclasses import dataclass
from typing import TYPE_CHECKING

from harpy.processing.types import MapTask, TransformTask, ReduceTask, FanoutTask
from harpy.dataset.base_classes import WriteType, ReadType, DataFragment
from harpy.session import Session
from harpy.quack import QuackContext
import pyarrow as pa
import pyarrow.parquet as pq
from pyarrow.dataset import Fragment
from uuid import uuid4
from harpy.processing.remote_exec import RemoteExecMetadata

if TYPE_CHECKING:
    from harpy.dataset.dataset import Dataset
    from harpy.dataset.dataset import ReadOptions
    from harpy.dataset.dataset import WriteOptions

def definition_quack_query_arrow(query: str) -> DataFragment:
    with QuackContext() as qc:
        return DataFragment(qc.sql(query).fetch_arrow_table(), 0)

# Redistribution tasks
def definition_quack_query_count(query: str) -> int:
    with QuackContext() as qc:
        return qc.sql(f"SELECT COUNT(*) FROM ({query})").fetchone()[0]

def definition_quack_query_arrow_distributed(row_count: int, query: str, fanout_count:int) -> DataFragment:
    remote_metadata = RemoteExecMetadata()
    fanout_index = int(remote_metadata.get_metadata("fanout_index"))
    # Calculate the limit and offsets
    limit = row_count // fanout_count
    offset = limit * fanout_index
    remainder = row_count % fanout_count
    if fanout_index == (fanout_count - 1):
        limit += remainder
    # Execute the query
    with QuackContext() as qc:
        return DataFragment(qc.sql(f"SELECT * FROM ({query}) LIMIT {limit} OFFSET {offset}").fetch_arrow_table(), fanout_index)

class SqlRead(ReadType):
    def __init__(self, dataset:"Dataset", query:str, options: "ReadOptions"):
        self.dataset = dataset
        self.read_options = options
        if query is None:
            raise ValueError("No query set")
        self._sql_ = query
        self.read_options.set_default("distribute_on_read", False)
        self.read_options.set_default("parallelism", 4)
    
    def __add_tasks__(self) -> None:
        if self.read_options.get_option("distribute_on_read"):
            # If we are distributing on read, we should use the parallelism
            parallelism = self.read_options.get_option("parallelism")
            map_count = MapTask(name="query_count", fun=definition_quack_query_count, kwargs={"query": self._sql_})
            fanout_task = FanoutTask(
                name="query_fanout", fun=definition_quack_query_arrow_distributed, kwargs={"query": self._sql_, "fanout_count": parallelism},
                fanout_count=parallelism
            )
            self.dataset._taskset_.add_maps([map_count])
            self.dataset._taskset_.add_fanout(fanout_task)
        else:
            # If we are not distributing on read, we should just run the query
            map_task = MapTask(name="query", fun=definition_quack_query_arrow, kwargs={"query": self._sql_})
            self.dataset._taskset_.add_maps([map_task])