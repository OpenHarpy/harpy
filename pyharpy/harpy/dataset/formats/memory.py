from dataclasses import dataclass
from typing import TYPE_CHECKING

from harpy.processing.types import MapTask, TransformTask, ReduceTask, FanoutTask
from harpy.dataset.base_classes import WriteType, ReadType, DataFragment
from harpy.session import Session
from harpy.quack import QuackContext
import pyarrow as pa
import pandas as pd
from uuid import uuid4
from harpy.processing.remote_exec import RemoteExecMetadata

if TYPE_CHECKING:
    from harpy.dataset.dataset import Dataset
    from harpy.dataset.dataset import ReadOptions
    from harpy.dataset.dataset import WriteOptions
    from harpy.tasksets import TaskSet

def read_pa_from_df(df: pd.DataFrame, index:int, n:int) -> DataFragment:
    if n == 1:
        return DataFragment(table=pa.Table.from_pandas(df), fragIndex=index)
    else:
        # Split the dataframe
        split_size = len(df) // n
        remainder = len(df) % n
        if index == n - 1:
            split_size += remainder
        return DataFragment(table=pa.Table.from_pandas(df.iloc[index*split_size:(index+1)*split_size]), fragIndex=index)

def read_pa_from_pa_table(table: pa.Table, index:int, n:int) -> DataFragment:
    if n == 1:
        return DataFragment(table=table, fragIndex=index)
    else:
        # Split the table
        split_size = table.num_rows // n
        remainder = table.num_rows % n
        if index == n - 1:
            split_size += remainder
        return DataFragment(table=table.slice(index*split_size, split_size), fragIndex=index)

def collect_to_pandas(*datafragments: DataFragment) -> pd.DataFrame:
    # Concatenate the tables
    tables = [df.table for df in datafragments]
    return pa.concat_tables(tables).to_pandas()

def collect_to_arrow(*datafragments: DataFragment) -> pa.Table:
    tables = [df.table for df in datafragments]
    return pa.concat_tables(tables)

class MemoryRead(ReadType):
    def __init__(self, df:pa.Table|pd.DataFrame, options: "ReadOptions"):
        self.read_options = options
        self.read_options.set_default("distribute_on_read", False)
        self.read_options.set_default("parallelism", 4)
        self._data_ = df

    def __add_tasks__(self, taskset: "TaskSet") -> None:
        selected_function = read_pa_from_df if isinstance(self._data_, pd.DataFrame) else read_pa_from_pa_table
        if self.read_options.get_option("distribute_on_read"):
            # If we are distributing on read, we should use the parallelism
            parallelism = self.read_options.get_option("parallelism")
            maps = [
                MapTask(
                    name="read_memory", fun=selected_function, kwargs={"df": self._data_, "index": i, "n": parallelism}
                )
                for i in range(parallelism)
            ]
        else:
            # If we are not distributing on read, we should just read the data
            taskset.add_maps(
                [
                    MapTask(
                        name="read_memory", fun=selected_function, kwargs={"df": self._data_, "index": 0, "n": 1}
                    )
                ]
           )
    
    def __repr__(self) -> str:
        return "MemoryRead(" + ("pandas" if isinstance(self._data_, pd.DataFrame) else "arrow") + ")"

class MemoryWrite(WriteType):
    def __init__(self, write_options: "WriteOptions"):
        self._write_options_ = write_options
        self._write_options_.set_default("write_engine", "pandas")
        self._write_options_.set_default("limit", None)
        if write_options.get_option("write_engine") not in ["pandas", "arrow", None]:
            raise ValueError("Unknown write engine")
        if write_options.get_option("limit") is not None:
            limit = write_options.get_option("limit")
            if type(limit) != int:
                raise ValueError("Limit must be an integer or None")
            if limit < 1:
                raise ValueError("Limit must be greater than 0 or None")


    def __add_tasks__(self, taskset: "TaskSet") -> None:
        if self._write_options_.get_option("write_engine") == "pandas":
            taskset.add_reduce(
                ReduceTask(name="to_pandas", fun=collect_to_pandas, limit=self._write_options_.get_option("limit"))
            )
        elif self._write_options_.get_option("write_engine") == "arrow":
            taskset.add_reduce(
                ReduceTask(name="to_arrow", fun=collect_to_arrow, limit=self._write_options_.get_option("limit"))
            )
        else:
            raise ValueError("Unknown write engine")
        
    def __repr__(self) -> str:
        if self._write_options_.get_option("limit") is not None:
            return "MemoryWrite(" + self._write_options_.get_option("write_engine") + ", limit=" + str(self._write_options_.get_option("limit")) + ")"
        return "MemoryWrite(" + self._write_options_.get_option("write_engine") + ")"
