// Function to parse query parameters from the URL
function getQueryParams() {
    const params = new URLSearchParams(window.location.search);
    return {
        start: params.get('start'),
        end: params.get('end'),
        fold: params.get('fold')
    };
}

// Function to calculate the last 3 months
function getLast3MonthsRange() {
    const now = new Date();
    const threeMonthsAgo = new Date();
    threeMonthsAgo.setMonth(now.getMonth() - 3);

    const formatDate = (date) => {
        const pad = (num) => String(num).padStart(2, '0');
        return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
    };

    return {
        start: formatDate(threeMonthsAgo),
        end: formatDate(now)
    };
}

// Function to set the date pickers' values
function setDatePickers() {
    const { start, end, fold } = getQueryParams();
    const defaults = getLast3MonthsRange();

    document.getElementById('start-date').value = start ? start.split('T')[0] : defaults.start;
    document.getElementById('end-date').value = end ? end.split('T')[0] : defaults.end;
    // Set the extra parameter in the hidden input
    if (fold) {
        document.getElementById('fold').value = fold;
    }
}

// Function to handle form submission
function handleFormSubmission(event) {
    const startDate = document.getElementById('start-date').value;
    const endDate = document.getElementById('end-date').value;

    // Create full timestamps
    const startDateTime = `${startDate}T00:00`;
    const endDateTime = `${endDate}T23:59`;

    // Set hidden input values
    document.getElementById('start-datetime').value = startDateTime;
    document.getElementById('end-datetime').value = endDateTime;
}

// Set date pickers on page load
window.addEventListener('DOMContentLoaded', setDatePickers);

// Add form submission handler
document.getElementById('date-form').addEventListener('submit', handleFormSubmission);