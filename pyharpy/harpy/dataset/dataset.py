from dataclasses import dataclass

from harpy.processing.types import MapTask, TransformTask, ReduceTask
from harpy.session import Session
from harpy.dataset.formats.parquet import ParquetRead, ParquetWrite

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
        reader = ParquetRead(self, path, self)
        self.dataset._reader_ = reader
        maps = reader.__get_maps__()
        self.dataset._taskset_.add_maps(maps)
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
        self.dataset._taskset_.add_transform(writer.__get_transform__())
        return self.dataset
    
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
        self._taskset_.add_transform(TransformTask(name="transform", fun=function, args=[], kwargs=kwargs))
        return self

    def add_reduce(self, function: callable, **kwargs) -> "Dataset":
        if self._sealed:
            raise ValueError("Dataset is sealed")
        self._taskset_.add_reduce(ReduceTask(name="reduce", fun=function, args=[], kwargs=kwargs))
        return self

    def execute(self):
        if not self._sealed:
            raise ValueError("Dataset is not sealed")
        return self._taskset_.run(collect=True, detailed=True)
    