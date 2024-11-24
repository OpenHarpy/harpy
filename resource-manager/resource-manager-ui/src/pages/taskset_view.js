var grouped = {};

const generateGraphTaskGroups = (graph, nodeId, elementId) => {
    if (!nodeId) return '';
    let edges = "";
    let dot = 'digraph G {';
    dot += 'compound=true;';
    const traverse = (currentNode, index) => {
        if (!currentNode) return '';

        let subgraph = `subgraph "cluster ${currentNode}" {`;
        subgraph += `label="${currentNode}";`;
        const node = graph[currentNode];
        if (index == 0) {
            subgraph += `"${currentNode}-out" [shape=box label="Output"];`;
        } else if (!node.next_node_id) {
            subgraph += `"${currentNode}-in" [shape=box label="Input"];`;
        } else {
            subgraph += `"${currentNode}-in" [shape=box label="Input"];`;
            subgraph += `"${currentNode}-out" [shape=box label="Output"];`;
        }

        if (node.tasks) {
            let tasks = '';
            for (const taskId in node.tasks) {
                tasks += `<tr><td>${node.tasks[taskId].name}</td></tr>`;
            }
            subgraph += `"${currentNode}-tasks" [
                        shape=box
                        label=<
                            <table BORDER="0" CELLBORDER="1">
                                ${tasks}
                            </table>
                        >
                    ];`;
        }
        if (node.runs) {
            let runs = '';
            for (const runId in node.runs) {
                runs += `<tr><td>${runId}[done=${node.runs[runId].is_done}/success=${node.runs[runId].success}]</td></tr>`;
            }
            subgraph += `"${currentNode}-runs" [
                        shape=box
                        label=<
                            <table BORDER="0" CELLBORDER="1">
                                ${runs}
                            </table>
                        >
                    ];`;
        }
        subgraph += '}';
        dot += subgraph;
        if (node.next_node_id) {
            edges += `"${currentNode}-out" -> "${node.next_node_id}-in";`;
        }
        traverse(node.next_node_id, index + 1);
    };
    traverse(nodeId, 0);
    dot += edges;
    dot += '}';

    const viz = new Viz();
    viz.renderSVGElement(dot)
        .then(element => {
            document.getElementById(elementId).innerHTML = '';
            document.getElementById(elementId).appendChild(element);
        })
        .catch(error => {
            console.error(error);
        });
};

const generatePlans = async (data, elementId) => {
    let shallowPlan = data.task_set_options['harpy.dataset.shallow_plan'];
    let plan = data.task_set_options['harpy.taskset.plan'];

    let planHtml = '';
    if (shallowPlan) {
        planHtml += `<pre>${shallowPlan}</pre><br>`;
    }
    if (plan) {
        planHtml += `<pre>Plan:</pre><pre>${plan}</pre>`;
    }

    document.getElementById(elementId).innerHTML = planHtml;
};

const loadEventForTaskset = async (event) => {
    // Generate and display the graph
    generateGraphTaskGroups(event.graph, event.task_group_root, 'graphview');
    //document.getElementById('graphview-run').innerHTML = graphHtmlRun;
    generatePlans(event, 'graphview-plans');
    // Add the taseset id to the title
    document.getElementById('event_id').innerText = `[${event.taskset_id}]`
    document.getElementById('taskset').hidden = false;
};

getQueryVariable = function (variable) {
    var query = window.location.search.substring(1);
    var vars = query.split("&");
    for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split("=");
        if (pair[0] == variable) {
            return pair[1];
        }
    }
    return (false);
};

const gid = getQueryVariable('event_group_id');
const currentHost = window.location.hostname;
const currentPort = window.location.port;
window.onload = function() {
    callResourceManagerEndpoint(`taskset_events/filter?event_group_id=` + gid)
        .then(data => {
            console.info('Loaded JSON:', data);
            decoded = data.events.map(datapoint => {
                data = JSON.parse(datapoint.EventLogJSON)
                data['event_time'] = new Date(datapoint.EventLogTime * 1000);
                return data
            });

            let index = 0;
            let numEvents = decoded.length;
            let startTime = new Date(0);
            let completedTime = new Date(0);
            let latestTime = new Date(0);

            // Events can be sorted by time and the task_set_progress as same time can have multiple events
            // We want to find the latest event that is either completed, failed or canceled. In case there are no such events, we will use the latest event

            // Find the latest event
            // We start by creating a hash map of the events, using the key as event_time and the value of task_set_progress
            let sortable = {};
            decoded.forEach((event, i) => {
                let key_1 = event.event_time.getTime();
                let key_2 = event.task_set_progress;
                let compound_key = key_1 + '_' + key_2;
                sortable[compound_key] = i;
            });

            // Custom sorting function to sort the events by time and prioritize completed, failed, and canceled events
            let sorted_events = Object.keys(sortable).sort((a, b) => {
                let [a_time, a_progress] = a.split('_');
                let [b_time, b_progress] = b.split('_');

                // Sort by time first
                if (a_time !== b_time) {
                    return a_time - b_time;
                }

                // Prioritize completed, failed, and canceled events to be last
                const priority = { 'completed': 3, 'failed': 2, 'canceled': 1 };
                return (priority[a_progress] || 0) - (priority[b_progress] || 0);
            });
            // Now we can find the latest event
            first_event = decoded[sortable[sorted_events[0]]];
            latest_event = decoded[sortable[sorted_events[sorted_events.length - 1]]];
            console.info('First event:', first_event);
            console.info('Latest event:', latest_event);

            is_completed = (latest_event.task_set_progress === 'completed' || latest_event.task_set_progress === 'failed' || latest_event.task_set_progress === 'canceled');
            console.info('Is completed:', is_completed, latest_event.task_set_progress);

            summary = {
                'numEvents': numEvents,
                'status': latest_event.task_set_progress,
                'result': latest_event.task_set_status,
                'startTime': first_event.event_time,
                'completedTime': (is_completed) ? latest_event.event_time : 'N/A',
                'latestTime': latest_event.event_time,
                'duration': (is_completed) ? toReadableDuration(latest_event.event_time - first_event.event_time) : 'N/A',
            };

            populateTableWithData([summary], 'summary', 'overview');
            loadEventForTaskset(latest_event);
        })
        .catch(error => console.error('Error loading JSON:', error));
}