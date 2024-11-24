// harpy_ui/main.js
const loadData = async () => {
    await populateTableWithDataFromAPI('config', 'configsTable', 'configs');
}

loadData();
// Add timer to refresh the tables every 5 seconds
setInterval(
    async () => {
        await loadData();
    },
    5000
);