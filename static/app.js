// websocket here?
const wsManager = new WebSocketManager();
wsManager.connect();

// for now terrible wait for connection while i figure out how to otherwise return once connected
setTimeout(function(){
    const evtMsg = new EventMessage("join_match", '{"time_control":"20m"}');
    console.log(JSON.stringify(evtMsg));
    wsManager.send(evtMsg);
}, 2000);

// start game
createBoard("light");

const allSquares = document.querySelectorAll(".square");

allSquares.forEach( square => {
    square.addEventListener('dragstart', dragStart);
    square.addEventListener('dragover', dragOver);
    square.addEventListener('drop', websocketDragDrop(wsManager));
})