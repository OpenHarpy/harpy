from dataclasses import dataclass
from typing import TYPE_CHECKING

from harpy.processing.types import MapTask, TransformTask, ReduceTask
from harpy.dataset.base_classes import WriteType, ReadType, DataFragment
from harpy.session import Session
import pyarrow as pa
import pyarrow.parquet as pq
from pyarrow.dataset import Fragment
from uuid import uuid4

if TYPE_CHECKING:
    from harpy.dataset.dataset import Dataset
    from harpy.dataset.dataset import ReadOptions
    from harpy.dataset.dataset import WriteOptions
    from harpy.tasksets import TaskSet


def read_pa_from_fragment(fragment: Fragment, index:int) -> DataFragment:
    return DataFragment(table=fragment.to_table(), fragIndex=index)

def write_pa_to_parquet(data_fragment: DataFragment, path:str, write_idx:str) -> None:
    parquet_name = path + "/part-" + write_idx +  '-' + str(data_fragment.fragIndex) + ".parquet"
    pq.write_table(data_fragment.table, parquet_name)

# We can submit an intermediate task to collect the fragments

def collect_frags(location: str) -> list[Fragment]:
    return pq.ParquetDataset(location).fragments

def collect_fragments(location: str) -> list[Fragment]:
    ts = Session().create_task_set()
    ts.add_maps(
        [
            MapTask(name="collect_fragments", fun=collect_frags, kwargs={"location": location})
        ]
    )
    result = ts.run(collect=True)[0]
    return result

class ParquetRead(ReadType):
    def __init__(self, location:str, options: "ReadOptions"):
        self.read_options = options
        if location is None:
            raise ValueError("No location set")
        self._parquet_path_ = location
    
    def __get_maps_fragment__(self) -> list[MapTask]:
        frags = collect_fragments(self._parquet_path_)
        return [
            MapTask(
                name="read_parquet", fun=read_pa_from_fragment, kwargs={"fragment":frag, "index": index }
            )
            for index, frag in enumerate(frags)
        ]
    
    def __add_tasks__(self, taskset: "TaskSet") -> None:
        maps = self.__get_maps_fragment__()
        taskset.add_maps(maps)
    
    def __repr__(self) -> str:
        return f"ParquetRead({self._parquet_path_})"

class ParquetWrite(WriteType):
    def __init__(self, write_options: "WriteOptions", location: str):
        self._parquet_path_ = location
        self._write_options_ = write_options
        self._write_options_.set_default("write_mode", "overwrite")

    def __add_tasks__(self, taskset: "TaskSet") -> None:
        # Parquet requires that the folder exists before writing
        if self._write_options_.get_option("write_mode") == "overwrite":
            Session().fs.rm(self._parquet_path_, recursive=True)
        Session().fs.mkdir(self._parquet_path_)
        transform_task = TransformTask(
            name="write_parquet", fun=write_pa_to_parquet, kwargs={ "path": self._parquet_path_, "write_idx": str(uuid4()) }
        )
        taskset.add_transform(transform_task)

    def __repr__(self) -> str:
        return f"ParquetWrite({self._parquet_path_})"