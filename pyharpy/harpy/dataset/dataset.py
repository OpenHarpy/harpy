from dataclasses import dataclass
import pandas as pd
import pyarrow as pa

from harpy.processing.types import MapTask, TransformTask, ReduceTask
from harpy.session import Session
from harpy.dataset.formats.parquet import ParquetRead, ParquetWrite
from harpy.dataset.formats.sql import SqlRead
from harpy.dataset.formats.memory import MemoryRead, MemoryWrite
from harpy.dataset.base_classes import DataFragment, TaskType, ReadType, WriteType

@dataclass
class ReadOptions:
    def __init__(self, dataset: "Dataset"):
        self.dataset = dataset
        self.options = {}
    
    def option(self, key:str, value:str) -> "ReadOptions":
        self.options[key] = value
        return self
    
    def set_default(self, key:str, value:str) -> "ReadOptions":
        if key not in self.options:
            self.options[key] = value
        return self
    def get_option(self, key:str) -> str:
        return self.options[key]

    def parquet(self, path: str) -> 'Dataset':
        reader = ParquetRead(path, self)
        self.dataset._reader_ = reader
        return self.dataset
    
    def sql(self, query: str) -> 'Dataset':
        reader = SqlRead(query, self)
        self.dataset._reader_ = reader
        return self.dataset
    
    def memory(self, df: pd.DataFrame|pa.Table) -> 'Dataset':
        reader = MemoryRead(df, self)
        self.dataset._reader_ = reader
        return self.dataset

class WriteOptions:
    def __init__(self, dataset: "Dataset"):
        self.dataset = dataset
        self.options = {}
    
    def option(self, key:str, value:str) -> "WriteOptions":
        self.options[key] = value
        return self
        
    def set_default(self, key:str, value:str) -> "ReadOptions":
        if key not in self.options:
            self.options[key] = value
        return self

    def get_option(self, key:str) -> str:
        return self.options.get(key, None)

    def parquet(self, path: str) -> 'Dataset':
        writer = ParquetWrite(self, path)
        self.dataset._writer_ = writer
        return self.dataset
    
    def memory(self) -> 'Dataset':
        writer = MemoryWrite(self)
        self.dataset._writer_ = writer
        return self.dataset

class ExecutableDataset:
    def __init__(self, read: ReadType, write: WriteType, tasks: list[TaskType]):
        self._reader_ = read
        self._writer_ = write
        self._tasks_ = tasks
        self._taskset_ = None

    def __make_taskset__(self):
        if self._reader_ is None:
            raise ValueError("No reader set")
        if self._taskset_ is None:
            self._taskset_ = Session().create_task_set()
            if self._reader_ is not None:
                self._reader_.__add_tasks__(self._taskset_)
            for transform in self._tasks_:
                transform.__add_tasks__(self._taskset_)
            if self._writer_ is not None:
                self._writer_.__add_tasks__(self._taskset_)

    def execute(self, collect: bool = False, detailed: bool = False):
        self.__make_taskset__()
        return self._taskset_.run(collect=collect, detailed=detailed)
    
    def explain(self, return_plan=False):
        self.__make_taskset__()
        return self._taskset_.explain(return_plan=return_plan)
    
    def __del__(self):
        if self._taskset_ is not None:
            self._taskset_.__dismantle__()

class Dataset:
    def __init__(self) -> "Dataset":
        self.read = ReadOptions(self)
        self.write = WriteOptions(self)
        self._reader_ = None
        self._transforms_ = []
        self._writer_ = None
    
    def add_task(self, transform: TaskType) -> "Dataset":
        if self.read is None:
            raise ValueError("Cannot add task: No reader set")
        if self._writer_ is not None:
            raise ValueError("Cannot add task: Writer already set")
        self._transforms_.append(transform)
        return self
    
    def execute(self, collect: bool = False, detailed: bool = False):
        if self._reader_ is None:
            raise ValueError("No reader set")
        if self._writer_ is None:
            raise ValueError("No writer set")
        exect_dataset = ExecutableDataset(self._reader_, self._writer_, self._transforms_)
        return exect_dataset.execute(collect=collect, detailed=detailed)
    
    def collect(self, detailed=False):
        if self._reader_ is None:
            raise ValueError("No reader set")
        if self._writer_ is None:
            raise ValueError("No writer set")
        exect_dataset = ExecutableDataset(self._reader_, self._writer_, self._transforms_)
        return exect_dataset.execute(collect=True, detailed=detailed)
    
    def show(self, limit_datafragments: int = 1):
        return self.to_pandas(limit_datafragments=limit_datafragments)
    
    def to_pandas(self, limit_datafragments: int = None) -> pd.DataFrame:
        self_clone = Dataset()
        self_clone._reader_ = self._reader_
        self_clone._transforms_ = self._transforms_
        self_clone.write.option('write_engine', 'pandas').option('limit', limit_datafragments).memory()
        return self_clone.collect()[0]
    
    def to_arrow(self, limit_datafragments: int = None) -> pa.Table:
        self_clone = Dataset()
        self_clone._reader_ = self._reader_
        self_clone._transforms_ = self._transforms_
        self_clone.write.option('write_engine', 'arrow').option('limit', limit_datafragments).memory()
        return self_clone.collect()[0]
    
    def explain(self, detailed=False, return_plan=False):
        lambda_add_layer = lambda strr, layer: " "+"--" * layer + f'> {strr}' + "\n"
        explain_str = "--- Dataset Plan ---\n"
        layer = 0
        if self._reader_ is not None:
            explain_str += lambda_add_layer(self._reader_.__repr__(), layer)
        layer += 1
        for transform in self._transforms_:
            explain_str += "  " * layer + transform.__repr__() + "\n"
            layer += 1
        if self._writer_ is not None:
            explain_str += lambda_add_layer(self._writer_.__repr__(), layer)
        
        if detailed:
            extra_plan = ExecutableDataset(self._reader_, self._writer_, self._transforms_).explain(return_plan=True)
            explain_str += "\n--- Taskset Plan ---\n"
            explain_str += extra_plan
        if return_plan:
            return explain_str
        print(explain_str)
