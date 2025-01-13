import pip
import importlib.metadata
from harpy.processing.types import MapTask, OneOffClusterTask
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from harpy.session import Session   

def install_package_task(package_names: list) -> int:
    return pip.main(['install', *package_names])

def uninstall_packages_task(package_names: list) -> int:
    return pip.main(['uninstall', *package_names])

def get_installed_packages_task() -> list:
    installed_packages = importlib.metadata.distributions()
    return sorted([f"{dist.metadata['Name']}=={dist.version}" for dist in installed_packages])

def install_packages(session:"Session", package_names: list) -> bool:
    ts = session.create_task_set(
        options={'harpy.taskset.name': 'pyharpy-install-package'}
    ).add_oneoff_cluster(
        OneOffClusterTask(name='install-packages', fun=install_package_task, kwargs={'package_names': package_names})
    )
    result = ts.run(collect=True)
    # Check if all the packages were installed
    if set(result) != {0}:
        return False
    return True

def uninstall_packages(session:"Session", package_names: list) -> bool:
    ts = session.create_task_set(
        options={'harpy.taskset.name': 'pyharpy-uninstall-package'}
    ).add_oneoff_cluster(
        OneOffClusterTask(name='uninstall-packages', fun=uninstall_packages_task, kwargs={'package_names': package_names})
    )
    result = ts.run(collect=True)
    if set(result) != {0}:
        return False
    return True

def get_installed_packages(session) -> list:
    ts = session.create_task_set().add_maps([MapTask(name='get-installed-packages', fun=get_installed_packages_task)])
    result = ts.run(collect=True)
    return result[0]