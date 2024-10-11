// harpy_ui/tableGenerator.js

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
