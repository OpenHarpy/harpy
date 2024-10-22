from uuid import uuid4
from IPython.display import display, HTML

class TerminalProgressViewer:
    def __init__(self):
        # Start a new progress viewer
        # In jupyter notebook you can use a dedicated output cell to display the progress
        # Create new output cell
        self.id = str(uuid4())
        self.display_handle = display(HTML('<div>Starting up...</div>'), display_id=self.id)
        self._steps = {}
        self._current_step = 0
    
    def add_step(self, step, task):
        if step not in self._steps:
            self._steps[step] = {}
        if task not in self._steps[step]:
            self._steps[step][task] = { 'progress': 'pending' }
    
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