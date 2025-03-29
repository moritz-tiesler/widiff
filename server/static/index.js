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
            const response = await fetch('http://localhost:8080/diff');
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
        const { diffstring, comment, user } = diffCache[timeframe];
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
                drawFileList: false
            }
        );
        diffUserFooter.textContent = `\u2014 ${user}`;
        diffCommentDiv.textContent = comment;
        diffCommentDiv.appendChild(diffUserFooter);
        diff2htmlUi.draw();

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

    // Initialize: Fetch all diffs
    fetchAllDiffs();
});