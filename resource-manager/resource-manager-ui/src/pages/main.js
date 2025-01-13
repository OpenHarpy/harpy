// harpy_ui/main.js

const getAddressIsAlive = async () => {
    try {
        const response = await fetch(resourceManagerURI+'health');
        return response.ok;   
    }
    catch (error) {
        return false;
    }
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