const loadEventForTaskset = async (group_id) => {
    let target = "/taskset_view.html?event_group_id=" + group_id;
    window.open(target, '_blank');
};

window.onload = function() {
    callResourceManagerEndpoint('taskset_events')
        .then(data => {
            decoded = data.events.map(datapoint => {
                data = JSON.parse(datapoint.EventLogJSON)
                data['event_time'] = new Date(datapoint.EventLogTime * 1000);
                data['group_id'] = datapoint.EventGroupID;
                return data
            });
            grouped = decoded.reduce((acc, obj) => {
                let key = obj.group_id;
                if (!acc[key]) {
                    acc[key] = [];
                }
                acc[key].push(obj);
                return acc;
            }, {});

            var tableData = [];
            const progressPriority = { 'completed': 1, 'cancelled': 2, 'running': 3 };
            const statusPriority = { 'success': 1, 'cancelled': 2, 'failed': 3, 'running': 4 };
            for (const [key, value] of Object.entries(grouped)) {
                let min_event_time = new Date(Math.min(...value.map(event => event.event_time)));
                let latest_event_time = new Date(Math.max(...value.map(event => event.event_time)));
                let taskset_name = value[0].task_set_options['harpy.taskset.name'];
                let group_id = value[0].group_id;
                let duration = toReadableDuration((latest_event_time - min_event_time));
                let event_count = value.length;
                let session_id = value[0].session_id;
                let taskset_status = value.map(event => event.task_set_status).sort((a, b) => statusPriority[a] - statusPriority[b])[0];
                let taskset_progress = value.map(event => event.task_set_progress).sort((a, b) => progressPriority[a] - progressPriority[b])[0];

                tableData.push({
                    taskset_id: `<div><span>${taskset_name}</span><br><a href='#${key}' id="${key}" onclick='loadEventForTaskset("${group_id}")' style="padding-top: 25px">` + key + "</a></div>",

                    min_event_time: min_event_time,
                    max_event_time: latest_event_time,
                    duration: duration,
                    taskset_status: taskset_status,
                    taskset_progress: taskset_progress,
                    event_count: event_count,
                    session_id: session_id
                });
            }

            tableData.sort((a, b) => b.max_event_time - a.max_event_time);
            tableData.forEach(row => {
                row.min_event_time = row.min_event_time.toISOString();
                row.max_event_time = row.max_event_time.toISOString();
            });
            populateTableWithData(tableData, 'tasksets', 'tasksets');
        })
        .catch(error => console.error('Error loading JSON:', error));
};