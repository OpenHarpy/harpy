from harpy.primitives import SingletonMeta

class RemoteExecMetadata(metaclass=SingletonMeta):
    def __init__(self):
        self._remote_exec_metadata = {}
    def add_metadata(self, key: str, value: str):
        self._remote_exec_metadata[key] = value
    def get_metadata(self, key: str):
        return self._remote_exec_metadata[key]
    def get_all_metadata(self):
        return self._remote_exec_metadata
    def is_running_remotely(self):
        return self.get_metadata('running_remotely') == 'True'