
from abc import ABC, abstractmethod
from dataclasses import dataclass

import pyarrow as pa
import pyarrow.parquet as pq
from pyarrow.dataset import Fragment
from harpy.session import Session
from harpy.processing.types import MapTask, ReduceTask, TransformTask
from harpy.tasksets import TaskSet

class WriteType(ABC):
    @abstractmethod
    def __add_tasks__(self, taskset: TaskSet):
        pass

    @abstractmethod
    def __repr__(self) -> str:
        pass

    def __str__(self) -> str:
        return self.__repr__()

class ReadType(ABC):  
    @abstractmethod
    def __add_tasks__(self, taskset: TaskSet):
        pass

    @abstractmethod
    def __repr__(self) -> str:
        pass

    def __str__(self) -> str:
        return self.__repr__()

class TaskType(ABC):
    @abstractmethod
    def __add_tasks__(self, taskset: TaskSet):
        pass
    
    @abstractmethod
    def __repr__(self) -> str:
        pass

    def __str__(self) -> str:
        return self.__repr__()

@dataclass
class DataFragment:
    table: pa.Table
    fragIndex: int
    