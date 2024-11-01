import pip
import importlib.metadata
from harpy.processing.types import MapTask

def install_package_task(package_names: list) -> int:
    return pip.main(['install', *package_names])

def uninstall_packages_task(package_names: list) -> int:
    return pip.main(['uninstall', *package_names])

def get_installed_packages_task() -> list:
    installed_packages = importlib.metadata.distributions()
    return sorted([f"{dist.metadata['Name']}=={dist.version}" for dist in installed_packages])

def install_packages(session, package_names: list) -> bool:
    ts = session.create_task_set().add_maps([MapTask(name='install-packages', fun=install_package_task, kwargs={'package_names': package_names})])
    result = ts.run(collect=True)
    if result[0] != 0:
        return False
    return True

def uninstall_packages(session, package_names: list) -> bool:
    ts = session.create_task_set().add_maps([MapTask(name='uninstall-packages', fun=uninstall_packages_task, kwargs={'package_names': package_names})])
    result = ts.run(collect=True)
    if result[0] != 0:
        return False
    return True

def get_installed_packages(session) -> list:
    ts = session.create_task_set().add_maps([MapTask(name='get-installed-packages', fun=get_installed_packages_task)])
    result = ts.run(collect=True)
    return result[0]