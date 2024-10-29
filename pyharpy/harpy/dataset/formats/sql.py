from dataclasses import dataclass
from typing import TYPE_CHECKING

from harpy.processing.types import MapTask, TransformTask, ReduceTask
from harpy.dataset.base_classes import WriteType, ReadType, DataFragment
from harpy.session import Session
from harpy.quack import QuackContext
import pyarrow as pa
import pyarrow.parquet as pq
from pyarrow.dataset import Fragment
from uuid import uuid4

if TYPE_CHECKING:
    from harpy.dataset.dataset import Dataset
    from harpy.dataset.dataset import ReadOptions
    from harpy.dataset.dataset import WriteOptions

def definition_quack_query_pandas(query: str) -> pa.Table:
    with QuackContext() as qc:
        return qc.sql(query).fetchdf()

class SqlRead(ReadType):
    def __init__(self, dataset:"Dataset", query:str, options: "ReadOptions"):
        self.dataset = dataset
        self.read_options = options
        if query is None:
            raise ValueError("No query set")
        self._sql_ = query
        self.read_options.set_default("map_strategies", "fragment")
    
    def __get_maps__(self) -> list[MapTask]:
        return [
            MapTask(
                name="quack_query_pandas", fun=definition_quack_query_pandas, args=[], kwargs={"query": self._sql_}
            )
        ]