// websocket here?
const gameManager = new GameManager();
gameManager.connect();

// for now terrible wait for connection while i figure out how to otherwise return once connected
setTimeout(function(){
    const evtMsg = new EventMessage("join_match", '{"time_control":"20m"}');
    gameManager.send(evtMsg);
}, 100);
