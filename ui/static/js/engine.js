const gameManager = new GameManager();

const queryString = window.location.search;
const urlParams = new URLSearchParams(queryString);
const elo = urlParams.get('elo');
const connectionMessage = new EventMessage("new_engine_match", `{"elo":${elo}}`);

gameManager.connect('/engines/ws', connectionMessage);