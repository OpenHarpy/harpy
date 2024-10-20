import os
import pandas as pd

from harpy.tasksets import TaskSet
from harpy.processing.types import MapTask, TaskSetResults
from harpy.exceptions.user_facing import FileSystemException

# FS functions
def fs_ls(path: str) -> pd.DataFrame:
    # We need to get the list of files in the directory and their sizes
    files = os.listdir(path)
    data = []
    for file in files:
        fullPath = f"{path}/{file}".replace("//", "/")
        data.append({
            "fileName": file,
            "filePath": fullPath,
            "sizeInBytes": os.path.getsize(fullPath)
        })
    return pd.DataFrame(data)

def fs_cat(path: str) -> str:
    with open(path, 'r') as file:
        return file.read()

def fs_head(path: str, n: int) -> str:
    with open(path, 'r') as file:
        return ''.join([file.readline() for _ in range(n)])

def fs_tail(path: str, n: int) -> str:
    with open(path, 'r') as file:
        lines = file.readlines()
        return ''.join(lines[-n:])

def fs_wc(path: str) -> int:
    with open(path, 'r') as file:
        return len(file.readlines())

def fs_mkdir(path: str, recursive: bool = False) -> bool:
    if recursive:
        os.makedirs(path, exist_ok=True)
    else:
        os.mkdir(path)
    return True

def fs_rm(path: str, recursive: bool = False) -> bool:
    if recursive:
        os.removedirs(path)
    else:
        os.remove(path)
    return True

def run_fs_command(session, func: callable, path: str, *args, **kwargs):
    ts: TaskSet = session.create_task_set()
    name = func.__name__
    mapper = MapTask(name=f"fs-command-{name}", fun=func, args=[], kwargs={"path": path, **kwargs})
    ts.add_maps([mapper])
    result: TaskSetResults = ts.execute()
    session.dismantle_taskset(ts)
    if result.success:
        return result.results[0].result
    else:
        raise FileSystemException(f"Filesystem command failed with error", error_text=result.results[0].std_err)

class FileSystem():
    def __init__(self, session):
        self._fs_functions = {
            "ls": lambda path: run_fs_command(session, fs_ls, path),
            "cat": lambda path: run_fs_command(session, fs_cat, path),
            "head": lambda path, n=10: run_fs_command(session, fs_head, path, n),
            "tail": lambda path, n=10: run_fs_command(session, fs_tail, path, n),
            "wc": lambda path: run_fs_command(session, fs_wc, path),
            "mkdir": lambda path, recursive=False: run_fs_command(session, fs_mkdir, path, recursive=recursive),
            "rm": lambda path, recursive=False: run_fs_command(session, fs_rm, path, recursive=recursive)
        }
    def __getattribute__(self, name: str) -> callable:
        if name in object.__getattribute__(self, "_fs_functions"):
            return object.__getattribute__(self, "_fs_functions")[name]
        else:
            return object.__getattribute__(self, name)