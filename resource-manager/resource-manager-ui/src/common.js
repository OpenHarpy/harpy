const thisHost = window.location.hostname;
const resourceManagerURI = 'http://'+thisHost+':8080/';

const callResourceManagerEndpoint = async (endpoint) => {
    const response = await fetch(`${resourceManagerURI}${endpoint}`);
    return response.json();
}

const populateTableWithData = async (data, tableId, under) => {
    const tableHtml = generateTable(data, tableId);
    document.getElementById(under).innerHTML = '';
    document.getElementById(under).innerHTML += tableHtml;
}

const populateTableWithDataFromAPI = async (endpoint, tableId, under) => {
    const data = await callResourceManagerEndpoint(endpoint);
    populateTableWithData(data, tableId, under);
}

const toReadableDuration = function (duration) {
    let milliseconds = duration % 1000;
    let seconds = Math.floor(duration / 1000);
    let minutes = Math.floor(seconds / 60);
    let hours = Math.floor(minutes / 60);
    let days = Math.floor(hours / 24);

    if (days > 0) {
        return `${days} days, ${hours % 24} hours, ${minutes % 60} minutes, ${seconds % 60} seconds${milliseconds > 0 ? `, ${milliseconds} milliseconds` : ''}`;
    } else if (hours > 0) {
        return `${hours} hours, ${minutes % 60} minutes, ${seconds % 60} seconds${milliseconds > 0 ? `, ${milliseconds} milliseconds` : ''}`;
    } else if (minutes > 0) {
        return `${minutes} minutes, ${seconds % 60} seconds${milliseconds > 0 ? `, ${milliseconds} milliseconds` : ''}`;
    } else if (seconds > 0) {
        return `${seconds} seconds${milliseconds > 0 ? `, ${milliseconds} milliseconds` : ''}`;
    } else {
        return `${milliseconds} milliseconds`;
    }
};

/* Navegation */
const goto = function (where) {
    if (where === 'home') {
        window.location = `/index.html`;
    } else if (where === 'taskset_view') {
        window.location = `/taskset_view.html`;
    } else if (where === 'tasksets') {
        window.location = `/tasksets.html`;
    } else if (where === 'configs') {
        window.location = `/configs.html`;
    }
}

// onload add the nav bar
document.addEventListener('DOMContentLoaded', async () => {
    const navBar = document.getElementById('nav-bar');
    navBar.innerHTML = `
        <ul>
            <li><a id="nav-dash" href="#" onclick="goto('home')">Dashboard</a></li>
            <li><a id="nav-configs" href="#" onclick="goto('configs')">Configs</a></li>
            <li><a id="nav-tasksets" href="#" onclick="goto('tasksets')">Tasksets</a></li>
        </ul>
    `;

    currentNav = window.location.pathname.split('/').pop().split('.')[0];
    if (currentNav === '' || currentNav === 'index') {
        currentNav = 'dash';
    }

    document.getElementById(`nav-${currentNav}`).classList.add('active');
    
});

/**
 * Transforms a JavaScript object into an HTML table and assigns an ID to the table.
 * @param {Object} data - The JavaScript object to transform.
 * @param {string} tableId - The ID to assign to the table.
 * @returns {string} - The HTML string of the table.
 */
function generateTable(data, tableId) {
    if (!Array.isArray(data) || data.length === 0) {
        return '<p>No data available</p>';
    }
    let table = `<table id="${tableId}">`;
    const keys = Object.keys(data[0]);

    // Generate table header
    table += '<thead><tr>';
    keys.forEach(key => {
        table += `<th>${key}</th>`;
    });
    table += '</tr></thead>';

    // Generate table body
    table += '<tbody>';
    data.forEach(row => {
        table += '<tr>';
        keys.forEach(key => {
            table += `<td>${row[key]}</td>`;
        });
        table += '</tr>';
    });
    table += '</tbody>';

    table += '</table>';
    return table;
}