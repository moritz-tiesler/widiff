document.addEventListener('DOMContentLoaded', function () {
    const timeframeSelect = document.getElementById('timeframe');
    const diffOutputDiv = document.getElementById('diff2html-output');
    const diffCommentDiv = document.getElementById('diff-comment');
    const diffUserFooter = document.getElementById('diff-user');
    const outputformatSelect = document.getElementById('output-format')
    const diffCache = {}; // Store fetched diffs

    // Fetch all diffs on page load
    async function fetchAllDiffs() {
        try {
            const response = await fetch('/diff');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();

            // Assuming the server returns JSON like this:
            // {
            //  "minute": "diff string for minute",
            //  "hour": "diff string for hour",
            //  "day": "diff string for day"
            // }

            diffCache.minute = data.minute || null; // Store diff in cache
            diffCache.hour = data.hour || null; // Store diff in cache
            diffCache.day = data.day || null; // Store diff in cache
        } catch (error) {
            console.error('Error fetching all diffs:', error);
            diffCache.minute = null;
            diffCache.hour = null;
            diffCache.day = null;
        }

        // Display initial diff after all fetches are complete
        displayDiff(timeframeSelect.value, outputformatSelect.value);
    }

    function formatComment(comment, user) {
        console.log(`${comment}\n\u2014${user}`)
        return `${comment}\n\u2014${user}`
    }

    function displayDiff(timeframe, format) {
        const { diffstring, comment, user, review } = diffCache[timeframe];
        if (diffstring === null) {
            diffOutputDiv.textContent = `Failed to load diff for ${timeframe}.`;
            return;
        }

        if (!diffstring) {
            diffOutputDiv.textContent = `No diff available for ${timeframe}.`;
            return;
        }

        const diff2htmlUi = new Diff2HtmlUI(
            diffOutputDiv,
            diffstring,
            {
                outputFormat: format, // Or 'line-by-line'
                synchronisedScroll: true,
                colorScheme: 'dark',
                highlight: false,
                fileListToggle: false,
                fileListStartVisible: false,
                fileContentToggle: false,
                drawFileList: true
            }
        );
        diffUserFooter.textContent = `\u2014 ${user}`;
        diffCommentDiv.textContent = review.replace(/\\n/g, '\n') ;
        diffCommentDiv.appendChild(diffUserFooter);
        diff2htmlUi.draw();

    }

    // use broadcast api to avoid opening extra connection on new tabs
    function initEventSource() {
        const evtSource = new EventSource('/notify')
        evtSource.onmessage = (event) => {
            update = JSON.parse(event.data);
            console.log(update)
            diffCache.minute = update.minute || null; // Store diff in cache
            diffCache.hour = update.hour || null; // Store diff in cache
            diffCache.day = update.day || null; // Store diff in cache
            displayDiff(timeframeSelect.value, outputformatSelect.value);
        }
        evtSource.onerror = (error) => {
            console.error("SSE error", error)
            console.dir(error)
            console.log(evtSource.readyState)
        }
    }

    // Listen for timeframe changes
    timeframeSelect.addEventListener('change', function () {
        const selectedTimeframe = timeframeSelect.value;
        displayDiff(selectedTimeframe, outputformatSelect.value); // Display diff from cache
    });

    outputformatSelect.addEventListener('change', () => {
        const selectedFormat = outputformatSelect.value;
        displayDiff(timeframeSelect.value, selectedFormat);
    })

    src = initEventSource();
    fetchAllDiffs();
});

async function fetchReview() {
    const button = document.querySelector('.fetch-button');
    button.classList.add('loading');

    try {
        // Simulate fetching data from an API
        const response = await simulateFetch('https://jsonplaceholder.typicode.com/todos/1');
        console.log('Data fetched:', response);
    } catch (error) {
        console.error('Error fetching data:', error);
    } finally {
        button.classList.remove('loading');
    }
}

function simulateFetch(url) {
    return fetch(url)
        .then(x => new Promise(
            resolve => setTimeout(
                () => resolve(x.json()),
                2000)
        ))
}