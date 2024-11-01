# TODO: Add consistency checks for the "instance_id" that is serving the session
from typing import List
import grpc
import atexit

from harpy.primitives import SingletonMeta, check_variable
from harpy.configs import Configs
from harpy.grpc_ce_protocol.ceprotocol_pb2 import (
    SessionRequest,
    SessionHandler,
    TaskSetHandler
)
from harpy.grpc_ce_protocol.ceprotocol_pb2_grpc import (
    SessionStub,
)
from harpy.tasksets import TaskSet
from harpy.processing.types import MapTask, TaskSetResults
from harpy.quack import QuackContext
from harpy.exceptions.user_facing import SQLException
from harpy.session.file_system import FileSystem
from harpy.session.block_read_write_proxy import BlockReadWriteProxy
from harpy.session.packages import install_packages, uninstall_packages, get_installed_packages

INVALID_SESSION_MESSAGE = "Session is not valid, please make sure to create a session before using it"
INVALID_SESSION_INSTANCE_REFRESHED_MESSAGE = "Session is not valid, it looks like the remote instance has been refreshed, please create a new session"

# The session needs a separate implementation of Singleton because we need to test if the session has requested a close
class SessionSingletonMeta(type):
    """
    This is a thread-safe implementation of Singleton.
    """
    _instances = {}

    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            instance = super().__call__(*args, **kwargs)
            cls._instances[cls] = instance
        else:
            # Check if the session has requested a close in this case we can create a new instance
            if cls._instances[cls]._close_requested:
                instance = super().__call__(*args, **kwargs)
                cls._instances[cls] = instance
        return cls._instances[cls]

def check_session():
    def decorator(func):
        def wrapper(*args, **kwargs):
            session_instance:Session = args[0]
            if session_instance._session_handler is None:
                raise ValueError(INVALID_SESSION_MESSAGE)
            if session_instance._instance_id != session_instance.__get_remote_instance_id__():
                raise ValueError(INVALID_SESSION_INSTANCE_REFRESHED_MESSAGE)
            return func(*args, **kwargs)
        return wrapper
    return decorator

def ensure_set_default(config_key, default_value):
    if Configs().get(config_key) is None:
        Configs().set(config_key, default_value)

class Session(metaclass=SessionSingletonMeta):
    def __init__(self):
        self.conf = Configs()
        self.__reset_instance__()
        # Register onexit function to close the session
        atexit.register(self.close)
        self.__init_session__()
    
    def __reset_instance__(self):
        ensure_set_default('harpy.sdk.remote.controller.grpcHost', 'localhost')
        ensure_set_default('harpy.sdk.remote.controller.grpcPort', '50051')
        ensure_set_default('harpy.tasks.node.request.type', 'small-4cpu-1gb')
        ensure_set_default('harpy.tasks.node.request.count', '1')
        self._close_requested: bool = False
        self._instance_id: str = None
        self._session_stub: SessionStub = None
        self._session_handler: SessionHandler = None
        self._controller_channel: grpc.Channel = None
        self._block_read_write_proxy: BlockReadWriteProxy = None
        self._session_tasksets: List[TaskSet] = []
        self._session_raw_tasksets_close_callbacks: List[callable] = []
        self._quack_ctx = QuackContext()
        self.fs = FileSystem(self)
    
    def __init_session__(self,) -> 'Session':
        # Setup the session
        if self._controller_channel is None:
            self._controller_channel = grpc.insecure_channel(f"{self.conf.get('harpy.sdk.remote.controller.grpcHost')}:{self.conf.get('harpy.sdk.remote.controller.grpcPort')}")
        if self._session_stub is None:
            self._session_stub = SessionStub(self._controller_channel)
            session_request: SessionRequest = SessionRequest(Options=self.conf.getAllConfigs())
            self._session_handler = self._session_stub.CreateSession(session_request)
            self._instance_id = self.__get_remote_instance_id__()
            return self
        else:
            return self

    def __get_remote_instance_id__(self,) -> 'str':
        if self._session_handler is None:
            return None
        instance_metadata = self._session_stub.GetInstanceID(self._session_handler)
        return instance_metadata.InstanceID

    @check_session()
    def get_raw_taskset_handler(self, raw_tasksets_close_callbacks) -> TaskSetHandler:
        taskSetHandler: TaskSetHandler = self._session_stub.CreateTaskSet(self._session_handler)
        self._session_raw_tasksets_close_callbacks.append(raw_tasksets_close_callbacks)
        return taskSetHandler
    
    @check_session()
    def install_packages(self, package_names: List[str]) -> int:
        return install_packages(self, package_names)
    
    @check_session()
    def uninstall_packages(self, package_names: List[str]) -> int:
        return uninstall_packages(self, package_names)
    
    @check_session()
    def get_installed_packages(self) -> List[str]:
        return get_installed_packages(self)
        
    @check_session()
    def get_session_id(self):
        return self._session_handler.sessionId
    
    @check_session()
    def get_session_handler(self) -> SessionHandler:
        return self._session_handler
    
    @check_session()
    def create_task_set(self) -> TaskSet:
        inst = TaskSet(self,)
        self._session_tasksets.append(inst)
        return inst
    
    @check_session()
    def get_block_read_write_proxy(self) -> BlockReadWriteProxy:
        if self._block_read_write_proxy is None:
            self._block_read_write_proxy = BlockReadWriteProxy(self._controller_channel, self._session_handler)
        return self._block_read_write_proxy
    
    def close(self):
        # Clear all the tasksets
        if self._instance_id != self.__get_remote_instance_id__():
            self.__reset_instance__()
            self._close_requested = True
            return
        if self._session_handler is None:
            return
        for taskset in self._session_tasksets:
            taskset.__dismantle__()
        for callback in self._session_raw_tasksets_close_callbacks:
            try:
                callback()
            except Exception as e:
                print(e) # TODO: We need to improve the logging here 
        # Close the block read write proxy
        if self._block_read_write_proxy is not None:
            self._block_read_write_proxy = None
        # Close the session
        result = self._session_stub.CloseSession(self._session_handler)
        if result.Success:
            self.__reset_instance__()
        self._close_requested = True
        return