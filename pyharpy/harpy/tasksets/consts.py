HTML_JUPYTER_PROGRESS_VIEWER="""
<style>
    .progress-viewer {
        display: row;
        flex-direction: column;
        align-items: center;
        width: 100%;
    }
    .progress-viewer .step {
        display: flex;
        flex-direction: column;
        align-items: left;
        width: 100%;
        max-width: 25%;
    }
    .progress-viewer .step .name {
        margin-bottom: 2px;
    }
    .progress-viewer .step .tasks {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        width: 100%;
    }
    .progress-viewer .step .counter {
        margin-left: 4px;
    }
    .progress-viewer .step .tasks .status {
        margin-top: auto;
        margin-bottom: auto;
        margin-left: 4px;
        width: 100%;
        max-width: 10px;
        min-width: 10px;
        height: 10px; /* Adjust height for thicker bars */
        background-color: grey;
        border-radius: 5px;
        position: relative;
        overflow: hidden;
    }
    .progress-viewer .step .status.in-progress {
        background: repeating-linear-gradient(
            127deg,
            rgb(79, 79, 238) 0,
            rgb(79, 79, 238) 5px,
            #1b1d83 0px,
            #1b1d83 10px
        );
        background-size: 200% 100%;
        animation: barbershop 10s linear infinite;
    }
    .progress-viewer .step .status.completed {
        background-color: green;
    }
    .progress-viewer .step .status.not-started {
        background-color: grey;
    }
    .progress-viewer .step .status.failed {
        background-color: red;
    }
    
    @keyframes barbershop {
        0% {
            background-position: 0 0;
        }
        100% {
            background-position: 200% 0;
        }
    }
</style>
"""