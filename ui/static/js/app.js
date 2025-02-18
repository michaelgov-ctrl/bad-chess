// websocket here?
const gameManager = new GameManager();
gameManager.connect();

// for now terrible wait for connection while i figure out how to otherwise return once connected
// really this should probably maybe be handled in the initial setup of the serveWS() of the
// websocket manager. TODO: look into that.
const queryString = window.location.search;
const urlParams = new URLSearchParams(queryString);
const timecontrol = urlParams.get('timecontrol');

setTimeout(function(){
    const evtMsg = new EventMessage("join_match", `{"time_control":"${timecontrol}"}`);
    gameManager.send(evtMsg);
    gameManager.interrupt()
        .catch((error) => {
            temporaryMessage(JSON.stringify(error));
        });
}, 100);
