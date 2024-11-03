import os
import shutil
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
    try:
        if recursive:
            os.makedirs(path, exist_ok=True)
        else:
            os.mkdir(path)
    except Exception as e:
        print(e)
        return False
    return True

def fs_rm(path: str, recursive: bool = False) -> bool:
    # Check if its root or . or ..
    if path in ["/", ".", ".."]:
        raise FileSystemException("Cannot delete root directory or . or ..")
    # We need to check if the path is a directory or a file
    if not os.path.exists(path):
        return False
    try:
        if recursive:
            shutil.rmtree(path)
        else:
            os.remove(path)
    except Exception as e:
        print(e)
        return False
    return True

def put(path: str, content: str) -> bool:
    with open(path, 'w') as file:
        file.write(content)
    return True
    

def put_binary(path: str, content: bytes) -> bool:
    with open(path, 'wb') as file:
        file.write(content)
    return True

def get(path: str) -> bytes:
    with open(path, 'rb') as file:
        return file.read()

def folder_exists(path: str) -> bool:
    return os.path.exists(path) and os.path.isdir(path)

def run_fs_command(session, func: callable, path: str, *args, **kwargs):
    ts: TaskSet = session.create_task_set()
    name = func.__name__
    mapper = MapTask(name=f"fs-command-{name}", fun=func, kwargs={"path": path, **kwargs})
    ts.add_maps([mapper])
    result: TaskSetResults = ts.collect(detailed=True)
    if result.success:
        return result.results[0].result
    else:
        raise FileSystemException(f"Filesystem command failed with error", error_text=result.results[0].std_err)

class FileSystem():
    def __init__(self, session):
        self._fs_functions = {
            "ls": lambda path: run_fs_command(session, fs_ls, path),
            "cat": lambda path: run_fs_command(session, fs_cat, path),
            "head": lambda path, n=10: run_fs_command(session, fs_head, path, n=n),
            "tail": lambda path, n=10: run_fs_command(session, fs_tail, path, n=n),
            "wc": lambda path: run_fs_command(session, fs_wc, path),
            "mkdir": lambda path, recursive=False: run_fs_command(session, fs_mkdir, path, recursive=recursive),
            "rm": lambda path, recursive=False: run_fs_command(session, fs_rm, path, recursive=recursive),
            "folder_exists": lambda path: run_fs_command(session, folder_exists, path),
            "put": lambda path, content: run_fs_command(session, put, path, content=content),
            "put_binary": lambda path, content: run_fs_command(session, put_binary, path, content=content),
            "get": lambda path: run_fs_command(session, get, path),
        }
    def __getattribute__(self, name: str) -> callable:
        if name in object.__getattribute__(self, "_fs_functions"):
            return object.__getattribute__(self, "_fs_functions")[name]
        else:
            return object.__getattribute__(self, name)