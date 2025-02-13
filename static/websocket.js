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
            case "propagate_move":
                console.log("received propagated move:", evtMsg.payload);
                //console.log("targeted piece if any:", e.target.parentNode.getAttribute("class"))
                const mv = algToMove(evtMsg.payload.player, evtMsg.payload.MoveEvent.move);
                console.log(mv);
                // mv contains that pawn goodness rn

                // TODO: update board with moved piece
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

class Move {
    startingSquare;
    movedPiece;
    targetSquare;
    targetPiece;
    capture;
    playerColor;

    constructor(sSq, mp, ts, tp, c, pc) {
        this.startingSquare = sSq;
        this.movedPiece = mp;
        this.targetSquare = ts;
        this.targetPiece = tp;
        this.capture = c;
        this.playerColor = pc;
    }

    moveToAlgebraicNotationString() {
        return this.movedPiece + (this.capture ? "x" : "") + this.targetSquare;
    }
}

/*
propagate move funny business
class Move {
    startingSquare;
    movedPiece;
    targetSquare;
    targetPiece;
    capture;
    playerColor;

    constructor(sSq, mp, ts, tp, c, pc) {
        this.startingSquare = sSq;
        this.movedPiece = mp;
        this.targetSquare = ts;
        this.targetPiece = tp;
        this.capture = c;
        this.playerColor = pc;
    }

    moveToAlgebraicNotationString() {
        return this.movedPiece + (this.capture ? "x" : "") + this.targetSquare;
    }

    findMovedPiece() {
        switch (this.movedPiece) {
            case "K":
                // TODO: reverse based on valid moves till the piece is found
                // mv.startingSquare = ;
                break;
            case "Q":
                // TODO: reverse based on valid moves till the piece is found
                // mv.startingSquare = ;
                break;
            case "R":
                // TODO: reverse based on valid moves till the piece is found
                // mv.startingSquare = ;
                break;
            case "B":
                // TODO: reverse based on valid moves till the piece is found
                // mv.startingSquare = ;
                break;
            case "N":
                // TODO: reverse based on valid moves till the piece is found
                // mv.startingSquare = ;
                break;
            default:
                // these are pawns my dudes
                if (this.playerColor === "light") {
                    if ( this.capture ) {
                        let check = document.querySelector(`[square-id="${this.targetSquare + width - 1}"]`)?.firstChild;
                        if ( check !== null ) {
                            this.movedPiece = check;
                            this.startingSquare = this.targetSquare - width - 1
                        } else {
                            this.movedPiece = document.querySelector(`[square-id="${this.targetSquare + width + 1}"]`)?.firstChild;
                            this.startingSquare = this.targetSquare - width + 1
                        }

                        return;                        
                    }

                    for (let i = 1; i < 3; i++) {
                        let check = document.querySelector(`[square-id="${this.targetSquare + (width * i)}"]`)?.firstChild;
                        if ( check !== null ) {
                            this.movedPiece = check
                            this.startingSquare = this.targetSquare + (width * i)

                            return;
                        }
                    }
                } else {
                    if ( this.capture ) {
                        let check = document.querySelector(`[square-id="${this.startingSquare - width - 1 }"]`)?.firstChild;
                        if ( check !== null ) {
                            this.movedPiece = check;
                            this.startingSquare = this.startingSquare - width - 1
                        } else {
                            this.movedPiece = document.querySelector(`[square-id="${this.startingSquare - width + 1}"]`)?.firstChild;
                            this.startingSquare = this.startingSquare - width + 1
                        }

                        return;
                    }

                    console.log("dark pawn was moved without capturing")
                    for (let i = 1; i < 3; i++) {
                        let check = document.querySelector(`[square-id="${this.targetSquare - (width * i)}"]`)?.firstChild;
                        console.log(check);
                        if ( check !== null ) {
                            this.movedPiece = check
                            this.startingSquare = this.targetSquare - (width * i)
                        
                            return;
                        }
                    }
                }

                break;
        }
    }
}

function algebraicNotationToSquareId(alg) {
    const file = Number(alg.charCodeAt(0) - 97);
    const rank = parseInt(alg.charAt(1), 10) - 1;
    // console.log("file", file, "rank", rank);
    return (rank * 8) - file;
}

function algToMove(color, alg) {
    console.log("player color:", color, "alg move:", alg);
    let mv = new Move(null, null, null, null, false, color);

    const pieceChars = ["K", "Q", "R", "B", "N"];
    if ( pieceChars.includes(alg.charAt(0)) ) {
        mv.movedPiece = alg.charAt(0);
        if ( alg.charAt(1) === "x" ) {
            mv.capture = true;
            mv.targetSquare = algebraicNotationToSquareId(alg.substring(2,4));
            mv.targetPiece = document.querySelector(`[square-id="${mv.targetSquare}"]`)?.firstChild;
        } else {
            mv.targetSquare = algebraicNotationToSquareId(alg.substring(1,3));
        }
    } else {
        mv.movedPiece = alg.charAt(0);
        if ( alg.charAt(1) === "x" ) {
            mv.capture = true;
            console.log(alg.substring(2,4))
            mv.targetSquare = algebraicNotationToSquareId(alg.substring(2,4));
            mv.targetPiece = document.querySelector(`[square-id="${mv.targetSquare}"]`)?.firstChild;
        } else { 
            mv.targetSquare = algebraicNotationToSquareId(alg.substring(0,2));
        }
    }

    mv.findMovedPiece();
    return mv

}
*/