__version__ = "0.1.0"

from harpy.session import Session
from harpy.processing.types import MapTask, ReduceTask, TransformTask, Result, TaskSetResults
from harpy.tasksets import TaskSet
from harpy.configs import Configs

def detect_envrion_type() -> str:
    try:
        from IPython import get_ipython
        if get_ipython() is not None:
            return 'jupyter'
    except ImportError:
        return 'cli'
    return 'cli'

Configs().set('harpy.sdk.client.env', detect_envrion_type())
