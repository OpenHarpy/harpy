from abc import ABC, abstractmethod
from uuid import uuid4
from IPython.display import display, HTML
from tqdm import tqdm
from harpy.configs import Configs
from threading import Timer
from harpy.tasksets.consts import HTML_JUPYTER_PROGRESS_VIEWER

from abc import ABC, abstractmethod

def refresh_after(func):
    def wrapper(self, *args, **kwargs):
        result = func(self, *args, **kwargs)
        self.__refresh_progress__()
        return result
    return wrapper

class ProgressViewer(ABC):
    def __init__(self):
        self._steps = {}
        self._current_step = 0

    @refresh_after
    def add_step(self, step, task):
        self.__add_step__(step, task)

    @refresh_after
    def add_progress(self, step, task):
        self.__add_step__(step, task)
    
    @refresh_after
    def task_running(self, step, task):
        self.__task_running__(step, task)
    
    @refresh_after
    def task_success(self, step, task):
        self.__task_success__(step, task)

    @refresh_after
    def task_fail(self, step, task):
        self.__task_fail__(step, task)
    
    @refresh_after
    def set_overall_status(self, overall_status):
        self.__set_overall_status__(overall_status)
    
    @refresh_after
    def set_context_info(self, context_info, context_working=False):
        self.__set_context_info__(context_info, context_working)

    @abstractmethod
    def __add_step__(self, step, task):
        pass

    @abstractmethod
    def __task_running__(self, step, task):
        pass

    @abstractmethod
    def __task_success__(self, step, task):
        pass

    @abstractmethod
    def __task_fail__(self, step, task):
        pass

    @abstractmethod
    def __refresh_progress__(self):
        pass
    
    @abstractmethod
    def __set_overall_status__(self, overall_status):
        pass
    
    @abstractmethod
    def __set_context_info__(self, context_info, context_working=False):
        pass

# Implementations for Jupyter and CLI
class JupyterProgressViewer(ProgressViewer):
    def __init__(self):
        super().__init__()
        self.id = str(uuid4())
        self.display_handle = display(HTML('<div>Starting up...</div>'), display_id=self.id)
        self.overall_status = None
        self._last_update_time = 0
        self.context_info = None
        self.context_working = False

    def __add_step__(self, step, task):
        if step not in self._steps:
            self._steps[step] = {}
        if task not in self._steps[step]:
            self._steps[step][task] = {'progress': 'not-started'}

    def __task_running__(self, step, task):
        self._steps[step][task]['progress'] = 'in-progress'

    def __task_success__(self, step, task):
        self._steps[step][task]['progress'] = 'completed'

    def __task_fail__(self, step, task):
        self._steps[step][task]['progress'] = 'failed'

    def __set_overall_status__(self, overall_status):
        self.overall_status = overall_status
        if overall_status == 'success':
            for step in self._steps:
                for task in self._steps[step]:
                    if self._steps[step][task]['progress'] == 'not-started':
                        self._steps[step][task]['progress'] = 'completed'
        elif overall_status == 'failed':
            for step in self._steps:
                for task in self._steps[step]:
                    if self._steps[step][task]['progress'] == 'not-started':
                        self._steps[step][task]['progress'] = 'failed'
    def __set_context_info__(self, context_info, context_working=False):
        self.context_info = context_info
        self.context_working = context_working

    def __refresh_progress__(self):
        """Updates the display in Jupyter with the new HTML."""
        html_content = self.__render_html__()
        self.display_handle.update(HTML(html_content))
    
    def __render_html__(self):
        """Generates the HTML for the progress viewer with `tasks` container."""
        html = HTML_JUPYTER_PROGRESS_VIEWER
        html += '<div class="progress-viewer">'
        if self.context_info:
            html += f'<div class="context-info">'
            if self.context_working:
                html += '<div class="context-working"></div>'
            html += f'<label>{self.context_info}</label></div>'
        for step, tasks in self._steps.items():
            html += f'<div class="step"><div class="name">{step}</div><div class="tasks">'
            for task, info in tasks.items():
                task_class = info['progress']
                html += f'<div class="status {task_class}"></div>'
            total_tasks = len(tasks)
            tasks_success = len([t for t in tasks.values() if t['progress'] == 'completed'])
            tasks_failed = len([t for t in tasks.values() if t['progress'] == 'failed'])
            html += '<div class="counter">'
            html += f'{tasks_success} / {total_tasks}' 
            if tasks_failed > 0:
                html += f' ({tasks_failed} failed)'
            html += '</div>'
            
            html += '</div></div>'  # Close tasks and step
        html += '</div>'  # Close progress-viewer
        return html

class CLIProgressViewer(ProgressViewer):
    def __init__(self):
        super().__init__()
        self.progress_bars = {}
        self.overall_status = None

    def __add_step__(self, step, task):
        if step not in self._steps:
            self._steps[step] = {}
        if task not in self._steps[step]:
            self._steps[step][task] = {'progress': 'pending'}
            if step not in self.progress_bars:
                self.progress_bars[step] = tqdm(total=1, desc=f'Step {step}', position=len(self.progress_bars))
        self.print_progress()

    def __task_running__(self, step, task):
        self._steps[step][task]['progress'] = 'running'
        self.print_progress()

    def __task_success__(self, step, task):
        self._steps[step][task]['progress'] = 'success'
        self.progress_bars[step].update(1 / len(self._steps[step]))
        self.print_progress()

    def __task_fail__(self, step, task):
        self._steps[step][task]['progress'] = 'fail'
        self.progress_bars[step].update(1 / len(self._steps[step]))
        self.print_progress()

    def __set_overall_status__(self, overall_status):
        self.overall_status = overall_status
        if overall_status == 'success':
            # Mark all pending tasks as success
            for step in self._steps:
                for task in self._steps[step]:
                    if self._steps[step][task]['progress'] == 'pending':
                        self._steps[step][task]['progress'] = 'success'
                        self.progress_bars[step].update(1 / len(self._steps[step]))
        self.print_progress()

    def __set_context_info__(self, context_info, context_working=False):
        pass

    def __refresh_progress__(self):
        for step, bar in self.progress_bars.items():
            bar.refresh()

def get_progress_viewer():
    if Configs().get('harpy.sdk.client.env') == 'jupyter':
        return JupyterProgressViewer()
    else:
        return CLIProgressViewer()