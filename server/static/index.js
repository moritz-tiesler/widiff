
// async function fetchDiffString() {
//     try {
//         const response = await fetch('http://localhost:8080/diff');
//         if (!response.ok) {
//             throw new Error(`HTTP error! status: ${response.status}`);
//         }
//         return response.text();
//     } catch (error) {
//         console.error('Error fetching diff:', error);
//         return null; // Or a default diff string, or handle the error differently
//     }
// }



// let currentDiff;
// function poll() {
//     fetchDiffString()
//         .then(d => {
//             currentDiff = d;
//             displayDiffView(currentDiff);
//             setTimeout(poll, 1000 * 60)
//         })
//         .catch(error => {
//             console.error('Error during poll, retrying...', error)
//             setTimeout(poll, 1000 * 5)
//         })
// }

// function displayDiffView(diff) {
//     var targetElement = document.getElementById('destination-elem-id');
//     var configuration = {
//         drawFileList: true,
//         fileListToggle: false,
//         fileListStartVisible: false,
//         fileContentToggle: false,
//         matching: 'lines',
//         highlight: false,
//         outputFormat: 'side-by-side',
//         synchronisedScroll: true,
//         renderNothingWhenEmpty: false,
//         colorScheme: 'dark',
//     };
//     var diff2htmlUi = new Diff2HtmlUI(targetElement, diff, configuration);
//     diff2htmlUi.draw();

//     const fileAnchor = document.querySelector('.d2h-file-name');
//     const pageName = fileAnchor.text;
//     fileAnchor.setAttribute('href', buildWikiLink(pageName));
//     fileAnchor.setAttribute('target', '_blank');
// }

// function buildWikiLink(pageName) {
//     const underscored = pageName.replace(' ', '_');
//     return `https://en.wikipedia.org/wiki/${underscored}`;
// }

// async function main() {
//     document.addEventListener('DOMContentLoaded', async () => {
//         const styleToggle = document.getElementById("styleToggle");
//         styleToggle.addEventListener("change", function () {
//             const elements = document.querySelectorAll('.d2h-code-line-ctn')
//             if (elements.length === 0) {
//                 return;
//             }
//             if (!this.checked) {
//                 for (e of elements) {
//                     e.style.whiteSpace = 'unset'; // Removes the background-color
//                     e.style.wordBreak = 'unset';        // Removes the border
//                 }
//             } else {
//                 for (e of elements) {
//                     e.style.whiteSpace = 'pre-wrap'; // Removes the background-color
//                     e.style.wordBreak = 'break-all';        // Removes the border
//                 }
//             }
//         });
//         diffString = await fetchDiffString();
//         currentDiff = diffString;
//         displayDiffView(currentDiff);
//         poll();
//     });
// }

// main();

document.addEventListener('DOMContentLoaded', function () {
    const timeframeSelect = document.getElementById('timeframe');
    const diffOutputDiv = document.getElementById('diff-output');
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
        displayDiff(timeframeSelect.value);
    }

    function displayDiff(timeframe) {
        const diffString = diffCache[timeframe];

        if (diffString === null) {
            diffOutputDiv.textContent = `Failed to load diff for ${timeframe}.`;
            return;
        }

        if (!diffString) {
            diffOutputDiv.textContent = `No diff available for ${timeframe}.`;
            return;
        }

        const diff2htmlUi = new Diff2HtmlUI(
            diffOutputDiv,
            diffString,
            {
                outputFormat: 'side-by-side', // Or 'line-by-line'
                synchronisedScroll: true,
                colorScheme: 'dark',
                highlight: false,
                fileListStartVisible: true,
                fileListToggle: false
            }
        );
        diff2htmlUi.draw();
    }

// Listen for timeframe changes
timeframeSelect.addEventListener('change', function () {
    const selectedTimeframe = timeframeSelect.value;
    displayDiff(selectedTimeframe); // Display diff from cache
});

// Initialize: Fetch all diffs
fetchAllDiffs();
});