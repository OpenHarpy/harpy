from abc import ABC, abstractmethod
from uuid import uuid4
from IPython.display import display, HTML
from tqdm import tqdm
from harpy.configs import Configs

class ProgressViewer(ABC):
    def __init__(self):
        self._steps = {}
        self._current_step = 0

    @abstractmethod
    def add_step(self, step, task):
        pass

    @abstractmethod
    def task_running(self, step, task):
        pass

    @abstractmethod
    def task_success(self, step, task):
        pass

    @abstractmethod
    def task_fail(self, step, task):
        pass

    @abstractmethod
    def print_progress(self):
        pass

class JupyterProgressViewer(ProgressViewer):
    def __init__(self):
        super().__init__()
        self.id = str(uuid4())
        self.display_handle = display(HTML('<div>Starting up...</div>'), display_id=self.id)

    def add_step(self, step, task):
        if step not in self._steps:
            self._steps[step] = {}
        if task not in self._steps[step]:
            self._steps[step][task] = {'progress': 'pending'}

    def task_running(self, step, task):
        self._steps[step][task]['progress'] = 'running'

    def task_success(self, step, task):
        self._steps[step][task]['progress'] = 'success'

    def task_fail(self, step, task):
        self._steps[step][task]['progress'] = 'fail'

    def print_progress(self):
        data = "<div>"
        for step in self._steps:
            num_tasks = len(self._steps[step])
            num_tasks_success = len([t for _, t in self._steps[step].items() if t['progress'] == 'success'])
            num_tasks_fail = len([t for _, t in self._steps[step].items() if t['progress'] == 'fail'])
            data += f'<div>Step {step}:</div>'
            data += "<div>"
            progress_bar = ''
            for _, t in self._steps[step].items():
                if t['progress'] == 'pending':
                    progress_bar += '<span style="display: inline-block; width: 10px; height: 10px; background-color: gray; border-radius: 25%; margin-right: 2px;"></span>'
                elif t['progress'] == 'running':
                    progress_bar += '<span style="display: inline-block; width: 10px; height: 10px; background-color: blue; border-radius: 25%; margin-right: 2px;"></span>'
                elif t['progress'] == 'success':
                    progress_bar += '<span style="display: inline-block; width: 10px; height: 10px; background-color: green; border-radius: 25%; margin-right: 2px;"></span>'
                elif t['progress'] == 'fail':
                    progress_bar += '<span style="display: inline-block; width: 10px; height: 10px; background-color: red; border-radius: 25%; margin-right: 2px;"></span>'
            data += progress_bar
            if num_tasks_fail > 0:
                data += f'<span style="padding:5px;">{num_tasks_success}/{num_tasks} ({num_tasks_fail} failed)</span>'
            data += f'<span style="padding:5px;">{num_tasks_success}/{num_tasks}</span>'
            data += "</div>"
        data += "</div>"
        self.display_handle.update(HTML(data))

class CLIProgressViewer(ProgressViewer):
    def __init__(self):
        super().__init__()
        self.progress_bars = {}

    def add_step(self, step, task):
        if step not in self._steps:
            self._steps[step] = {}
        if task not in self._steps[step]:
            self._steps[step][task] = {'progress': 'pending'}
            if step not in self.progress_bars:
                self.progress_bars[step] = tqdm(total=1, desc=f'Step {step}', position=len(self.progress_bars))

    def task_running(self, step, task):
        self._steps[step][task]['progress'] = 'running'

    def task_success(self, step, task):
        self._steps[step][task]['progress'] = 'success'
        self.progress_bars[step].update(1 / len(self._steps[step]))

    def task_fail(self, step, task):
        self._steps[step][task]['progress'] = 'fail'
        self.progress_bars[step].update(1 / len(self._steps[step]))

    def print_progress(self):
        for step, bar in self.progress_bars.items():
            bar.refresh()

def get_progress_viewer():
    if Configs().get('harpy.sdk.client.env') == 'jupyter':
        return JupyterProgressViewer()
    else:
        return CLIProgressViewer()