// harpy_ui/main.js

document.addEventListener('DOMContentLoaded', function() {
    const sections = document.querySelectorAll('section');
    const navLinks = document.querySelectorAll('nav ul li a');
    const defaultSection = 'dashboard';

    function showSection(event) {
        event.preventDefault();
        const targetId = event.target.getAttribute('href').substring(1);

        sections.forEach(section => {
            if (section.id === targetId) {
                section.style.display = 'block';
            } else {
                section.style.display = 'none';
            }
        });

        navLinks.forEach(link => {
            if (link.getAttribute('href').substring(1) === targetId) {
                link.classList.add('active');
            } else {
                link.classList.remove('active');
            }
        });
    }

    navLinks.forEach(link => {
        link.addEventListener('click', showSection);
    });

    // Initially show the home section and hide others
    sections.forEach((section, index) => {
        section.style.display = section.id === 'dashboard' ? 'block' : 'none';
    });

    navLinks.forEach((link, index) => {
        link.classList.remove('active');
        if (link.getAttribute('href').substring(1) === defaultSection) {
            link.classList.add('active');
        }
    });
});

const thisHost = window.location.hostname;
const resourceManagerURI = 'http://'+thisHost+':8080/';
const getAddressIsAlive = async () => {
    try {
        const response = await fetch(resourceManagerURI+'health');
        return response.ok;   
    }
    catch (error) {
        return false;
    }
}
    
const callResourceManagerEndpoint = async (endpoint) => {
    const response = await fetch(`${resourceManagerURI}${endpoint}`);
    return response.json();
}

const populateTableWithDataFromAPI = async (endpoint, tableId, under) => {
    const data = await callResourceManagerEndpoint(endpoint);
    const tableHtml = generateTable(data, tableId);
    document.getElementById(under).innerHTML = '';
    document.getElementById(under).innerHTML += tableHtml;
}

const setContentVisibility = (visible) => {
    const sections = document.querySelectorAll('#content');
    sections.forEach(section => {
        section.style.display = visible ? 'block' : 'none';
    });
}

const refreshStatus = async (newStatus) => {
    now = new Date();
    document.getElementById('status').innerHTML = `Resource Manager is ${newStatus}`;
    document.getElementById('last-update').innerHTML = `Last updated: ${now.toLocaleTimeString()}`;
}
const loadData = async () => {
    const healthCheck = await getAddressIsAlive();
    // If the health check is successful, load the data
    if (healthCheck) {
        refreshStatus('available');
        setContentVisibility(true);
        await populateTableWithDataFromAPI('nodes', 'nodesTable', 'nodes');
        await populateTableWithDataFromAPI('requests', 'requestsTable', 'resource-requests');
        await populateTableWithDataFromAPI('config', 'configsTable', 'configs');
    } else {
        refreshStatus('unavailable');
        setContentVisibility(false);
    }
}

loadData();
// Add timer to refresh the tables every 5 seconds
setInterval(
    async () => {
        await loadData();
    },
    5000
);