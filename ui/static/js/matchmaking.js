const gameManager = new GameManager();

const queryString = window.location.search;
const urlParams = new URLSearchParams(queryString);
const timecontrol = urlParams.get('timecontrol');
const connectionMessage = new EventMessage("join_match", `{"time_control":"${timecontrol}"}`);

gameManager.connect('/matches/ws', connectionMessage);