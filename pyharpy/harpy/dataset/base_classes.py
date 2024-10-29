
from abc import ABC, abstractmethod
from dataclasses import dataclass

import pyarrow as pa
import pyarrow.parquet as pq
from pyarrow.dataset import Fragment
from harpy.session import Session
from harpy.processing.types import MapTask, ReduceTask, TransformTask

class WriteType(ABC):
    @abstractmethod
    def write(self):
        pass

    @abstractmethod
    def __get_transform__(self):
        pass

class ReadType(ABC):
    def __init__(self):
        self._options_ = {}
    
    @abstractmethod
    def __get_maps__(self):
        pass

@dataclass
class DataFragment:
    table: pa.Table
    fragIndex: int
    