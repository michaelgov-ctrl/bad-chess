class EventMessage {
    type;
    payload;

    constructor(type, payload) {
        this.type = type;
        this.payload = payload;
    }

    toJSON() {
        return {
            type: this.type,
            //payload: this.payload,
            payload: JSON.parse(this.payload),
        };
    }
}

class JoinMatchEventMessage {
    timeControl;

    constructor(timeControl) {
        this.timeControl = timeControl;
    }
}

class MakeMoveEventMessage {
    move;

    constructor(move) {
        this.move = move;
    }
}

class PropagateMoveEventMessage {
    playerColor;
    moveEventMessage;

    constructor(playerColor, mvEvtMsg) {
        this.playerColor = playerColor;
        this.moveEventMessage = mvEvtMsg;
    }
}

class PropagatePositionEventMessage {
    playerColor;
    moveEventMessage;

    constructor(playerColor, mvEvtMsg) {
        this.playerColor = playerColor;
        this.moveEventMessage = mvEvtMsg;
    }
}

class ErrorEventMessage {
    error;
    
    constructor(err) {
        this.error = err;
    }
}

// TODO: cast received messages to appropriate class

function NewMatch(assignedEvtMsg) {
    const matchId = assignedEvtMsg.payload?.match_id;
    const player = assignedEvtMsg.payload?.pieces;
    // if ( matchId === null || player === null ) {
    if ( !matchId || !player ) {
        throw new Error("things are not working out for the old liz lemon")
    }

    matchInfoDisplay.textContent = "Match ID: " + matchId;
    createBoard(player);

    const allSquares = document.querySelectorAll(".square");
    allSquares.forEach( square => {
        square.addEventListener('dragstart', dragStart);
        square.addEventListener('dragover', dragOver);
        square.addEventListener('drop', websocketDragDrop(gameManager));
    })

    const queryString = window.location.search;
    const urlParams = new URLSearchParams(queryString);
    const timecontrol = urlParams.get('timecontrol');
    
    playerClock.textContent = timecontrol;
    opponentClock.textContent = timecontrol;
}

function HandlePositionPropagation(propagationEvtMsg) {
    const fen = propagationEvtMsg.payload?.fen;
    //if ( fen === null ) {
    if ( !fen ) {
        throw new Error("something went awry");
    }

    const currentPosition = fen.substring(0, fen.indexOf(' '));
    let squareId = 0;
    for (const c of currentPosition) {
        if ( c == '/' ) {
            continue;
        }

        if (c >= '0' && c <= '8') {
            for ( let i = 0; i < parseInt(c); i++ ) {
                const square = document.querySelector(`[square-id="${squareId}"]`).innerHTML = '';
                squareId++;
            }
            continue;
        }

        const square = document.querySelector(`[square-id="${squareId}"]`);
        square.innerHTML = fenCharToPiece(c);
        square.firstChild.setAttribute('draggable', true)
        squareId++;
    }

    changePlayer();
}

function HandleClockUpdate(clockUpdateEvtMsg) {
    const clockOwner = clockUpdateEvtMsg.payload?.clock_owner;
    const timeRemaining = clockUpdateEvtMsg.payload?.time_remaining;
    if ( !clockOwner || !timeRemaining ) {
        throw new Error("good god lemon");
    }

    const truncatedTime = timeRemaining.substring(0, timeRemaining.indexOf("."));
    if ( playerPieces === clockOwner ) {
        playerClock.textContent = truncatedTime + "s";
    } else {
        opponentClock.textContent = truncatedTime + "s";
    }
}

class GameManager {
    socket = null;
    interuptMessage = null;

    connect(endpoint, connMsg) {
        this.socket = new WebSocket('wss://bad-chess.com' + endpoint);

        this.socket.addEventListener('open', () => {
            console.log('ws conn opened');

            setTimeout(function(){
                gameManager.send(connMsg);
                gameManager.interrupt()
                    .catch((error) => {
                        temporaryMessage(JSON.stringify(error));
                    });
            }, 1000);
        });

        this.socket.addEventListener('message', (evt) => {
            const msg = JSON.parse(evt.data);
            this.routeEventMessage(msg);
        });

        this.socket.addEventListener('error', (err) => {
            console.error('resp error:', err);
        });

        this.socket.addEventListener('close', (c) => {
            console.log("ws conn closed", c)
            matchInfoDisplay.textContent = "match over";
            turnDisplay.textContent = "";
            this.socket = null;
        });
    }

    send(evtMsg) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(evtMsg));
        }
        else {
            console.log('cannot send message websocket not open.');
        }
    }

    interrupt() {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                if ( !this.interuptMessage ) {    
                    resolve("success");
                } else {
                    const msg = this.interuptMessage;
                    this.interuptMessage = null;
                    reject(msg);
                }
            }, 200)
        });
    }

    routeEventMessage(evtMsg) {
        if (evtMsg.type === undefined) {
            alert('no type field in the event');
        }
    
        switch (evtMsg.type) {
            case "assigned_match":
                try {
                    NewMatch(evtMsg);
                } catch (error) {
                    this.interuptMessage = error
                }
                
                break;
            case "propagate_position":
                try {
                    HandlePositionPropagation(evtMsg);
                } catch (error) {
                    this.interuptMessage = error;
                }

                break;
            case "match_over":
                matchInfoDisplay.textContent = "match over";
                turnDisplay.textContent = "";
                this.socket.close(1000, 'User initiated closure');
                break;
            case "match_error":
                this.interuptMessage = evtMsg.payload;
                
                break;
            case "clock_update":
                try {
                    HandleClockUpdate(evtMsg);
                } catch (error) {
                    this.interuptMessage = error;
                }

                break;
            default:
                break;
        }
    }
}
