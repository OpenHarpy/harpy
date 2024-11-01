from dataclasses import dataclass
import pandas as pd
import pyarrow as pa

from harpy.processing.types import MapTask, TransformTask, ReduceTask
from harpy.session import Session
from harpy.dataset.formats.parquet import ParquetRead, ParquetWrite
from harpy.dataset.formats.sql import SqlRead
from harpy.dataset.base_classes import DataFragment

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
        reader = ParquetRead(self.dataset, path, self)
        self.dataset._reader_ = reader
        reader.__add_tasks__()
        return self.dataset
    
    def sql(self, query: str) -> 'Dataset':
        reader = SqlRead(self.dataset, query, self)
        self.dataset._reader_ = reader
        reader.__add_tasks__()
        return self.dataset

class WriteOptions:
    def __init__(self, dataset: "Dataset"):
        self.dataset = dataset
        self.options = {}
    
    def option(self, key:str, value:str) -> "WriteOptions":
        self.dataset._writer_.options[key] = value
        return self
        
    def set_default(self, key:str, value:str) -> "ReadOptions":
        if key not in self.options:
            self.options[key] = value
        return self

    def get_option(self, key:str) -> str:
        return self.options[key]

    def parquet(self, path: str) -> 'Dataset':
        writer = ParquetWrite(self.dataset, self, path)
        self.dataset._writer_ = writer
        self.dataset._sealed = True
        writer.__add_tasks__()
        return self.dataset

def collect_to_pandas(*datafragments: DataFragment) -> pd.DataFrame:
    # Concatenate the tables
    tables = [df.table for df in datafragments]
    return pa.concat_tables(tables).to_pandas()

def collect_to_arrow(*datafragments: DataFragment) -> pa.Table:
    tables = [df.table for df in datafragments]
    return pa.concat_tables(tables)

class Dataset:
    def __init__(self) -> "Dataset":
        self.read = ReadOptions(self)
        self.write = WriteOptions(self)
        self._taskset_ = Session().create_task_set()
        self._sealed = False
        self._reader_ = None
        self._writer_ = None
    
    def add_transform(self, function: callable, **kwargs) -> "Dataset":
        if self._sealed:
            raise ValueError("Dataset is sealed")
        self._taskset_.add_transform(TransformTask(name="transform", fun=function, kwargs=kwargs))
        return self

    def add_reduce(self, function: callable, **kwargs) -> "Dataset":
        if self._sealed:
            raise ValueError("Dataset is sealed")
        self._taskset_.add_reduce(ReduceTask(name="reduce", fun=function, kwargs=kwargs))
        return self

    def execute(self):
        if not self._sealed:
            raise ValueError("Dataset is not sealed")
        return self._taskset_.run(collect=True, detailed=True)
    
    def show(self):
        if self._sealed:
            raise ValueError("Dataset is sealed")
    
    def collect(self, detailed=False):
        if self._sealed:
            raise ValueError("Dataset is sealed")
        return self._taskset_.run(collect=True, detailed=detailed)
    
    def show(self):
        return self.to_pandas(limit_datafragments=1)
    
    def to_pandas(self, limit_datafragments: int = None) -> pd.DataFrame:
        if self._sealed:
            raise ValueError("Dataset is sealed")
        self._taskset_.add_reduce(
            ReduceTask(name="to_pandas", fun=collect_to_pandas, limit=limit_datafragments)
        )
        return self._taskset_.run(collect=True)[0]
    
    def to_arrow(self, limit_datafragments: int = None) -> pa.Table:
        if self._sealed:
            raise ValueError("Dataset is sealed")
        self._taskset_.add_reduce(
            ReduceTask(name="to_arrow", fun=collect_to_arrow, limit=limit_datafragments)
        )
        return self._taskset_.run(collect=True)[0]
    
    def explain(self):
        return self._taskset_.explain()