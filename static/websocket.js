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

class Move {
    movedPiece;
    targetSquare;
    capture;
    playerColor;

    constructor(mp, ts, c, pc) {
        this.movedPiece = mp;
        this.targetSquare = ts;
        this.capture = c;
        this.playerColor = pc;
    }

    moveToAlgebraicNotationString() {
        return this.movedPiece + (this.capture ? "x" : "") + this.targetSquare;
    }
}


class WebSocketManager {
    socket = null;
    interuptMessage = null;

    connect() {
        this.socket = new WebSocket('ws://localhost:8080/ws');

        this.socket.addEventListener('open', () => {
            console.log('ws conn opened');
        });

        this.socket.addEventListener('message', (evt) => {
            const msg = JSON.parse(evt.data);
            this.routeEventMessage(msg);
        });

        this.socket.addEventListener('error', (err) => {
            console.error('resp error:', err);
        });

        this.socket.addEventListener('close', (c) => {
            console.log('ws conn closed', c);
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
                if ( this.interuptMessage === null ) {
                    resolve("success");
                } else {
                    const msg = this.interuptMessage;
                    this.interuptMessage = null;
                    console.log(msg, this.interuptMessage);
                    reject(msg);
                }
            }, 100)
        });
    }

    routeEventMessage(evtMsg) {
        if (evtMsg.type === undefined) {
            alert('no type field in the event');
        }
    
        switch (evtMsg.type) {
            case "propagate_position":
                console.log("received propagated position:", evtMsg.payload);
                const fen = evtMsg.payload?.fen;
                if ( fen === null ) {
                    alert("something went awry");
                    return;
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
                break;
            case "match_error":
                this.interuptMessage = evtMsg.payload;
                break;
            default:
                console.log('resp:', evtMsg);
                break;
        }
    }
}
