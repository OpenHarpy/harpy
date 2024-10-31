from abc import ABC, abstractmethod
from uuid import uuid4
from IPython.display import display, HTML
from tqdm import tqdm
from harpy.configs import Configs
from threading import Timer
from harpy.tasksets.consts import HTML_JUPYTER_PROGRESS_VIEWER

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
    
    @abstractmethod
    def set_overall_status(self, overall_status):
        pass

class JupyterProgressViewer(ProgressViewer):
    def __init__(self):
        super().__init__()
        self.id = str(uuid4())
        self.display_handle = display(HTML('<div>Starting up...</div>'), display_id=self.id)
        self.overall_status = None
        self._last_update_time = 0
        self._debounce_timer = None

    def add_step(self, step, task):
        if step not in self._steps:
            self._steps[step] = {}
        if task not in self._steps[step]:
            self._steps[step][task] = {'progress': 'not-started'}
        self.print_progress()

    def task_running(self, step, task):
        self._steps[step][task]['progress'] = 'in-progress'
        self.print_progress()

    def task_success(self, step, task):
        self._steps[step][task]['progress'] = 'completed'
        self.print_progress()

    def task_fail(self, step, task):
        self._steps[step][task]['progress'] = 'failed'
        self.print_progress()

    def set_overall_status(self, overall_status):
        self.overall_status = overall_status
        if overall_status == 'success':
            for step in self._steps:
                for task in self._steps[step]:
                    if self._steps[step][task]['progress'] == 'not-started':
                        self._steps[step][task]['progress'] = 'completed'
        self.print_progress()

    def render_html(self):
        """Generates the HTML for the progress viewer with `tasks` container."""
        html = HTML_JUPYTER_PROGRESS_VIEWER
        html += '<div class="progress-viewer">'
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

    def print_progress(self):
        """Updates the display in Jupyter with the new HTML."""
        html_content = self.render_html()
        self.display_handle.update(HTML(html_content))

class CLIProgressViewer(ProgressViewer):
    def __init__(self):
        super().__init__()
        self.progress_bars = {}
        self.overall_status = None

    def add_step(self, step, task):
        if step not in self._steps:
            self._steps[step] = {}
        if task not in self._steps[step]:
            self._steps[step][task] = {'progress': 'pending'}
            if step not in self.progress_bars:
                self.progress_bars[step] = tqdm(total=1, desc=f'Step {step}', position=len(self.progress_bars))
        self.print_progress()

    def task_running(self, step, task):
        self._steps[step][task]['progress'] = 'running'
        self.print_progress()

    def task_success(self, step, task):
        self._steps[step][task]['progress'] = 'success'
        self.progress_bars[step].update(1 / len(self._steps[step]))
        self.print_progress()

    def task_fail(self, step, task):
        self._steps[step][task]['progress'] = 'fail'
        self.progress_bars[step].update(1 / len(self._steps[step]))
        self.print_progress()

    def set_overall_status(self, overall_status):
        self.overall_status = overall_status
        if overall_status == 'success':
            # Mark all pending tasks as success
            for step in self._steps:
                for task in self._steps[step]:
                    if self._steps[step][task]['progress'] == 'pending':
                        self._steps[step][task]['progress'] = 'success'
                        self.progress_bars[step].update(1 / len(self._steps[step]))
        self.print_progress()

    def print_progress(self):
        for step, bar in self.progress_bars.items():
            bar.refresh()

def get_progress_viewer():
    if Configs().get('harpy.sdk.client.env') == 'jupyter':
        return JupyterProgressViewer()
    else:
        return CLIProgressViewer()