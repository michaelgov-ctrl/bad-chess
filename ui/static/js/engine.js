document.addEventListener("DOMContentLoaded", function() {
    var footerElement = document.getElementsByTagName("footer")[0];
    if (footerElement) {
        footerElement.innerHTML = "You're playing " + "<a href='https://github.com/official-stockfish/Stockfish' target='_blank'>Stockfish</a>";
    }
});

const gameManager = new GameManager();

const queryString = window.location.search;
const urlParams = new URLSearchParams(queryString);
const elo = urlParams.get('elo');
const connectionMessage = new EventMessage("new_engine_match", `{"elo":${elo}}`);

gameManager.connect('/engines/ws', connectionMessage);