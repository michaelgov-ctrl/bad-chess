import * as readline from 'readline';

class EventMessage {
    type: string;
    payload: string;

    constructor(type, payload){
        this.type = type;
        this.payload = payload;
    }

    toJSON(): object {
        return {
            type: this.type,
            //payload: this.payload,
            payload: JSON.parse(this.payload),
        };
    }
}

class JoinMatchEventMessage {
    timeControl: string;

    constructor(timeControl){
        this.timeControl = timeControl;
    }
}

class MakeMoveEventMessage {
    move: string;

    constructor(move){
        this.move = move;
    }
}

class PropagateMoveEventMessage {
    playerColor: string;
    moveEventMessage: MakeMoveEventMessage;

    constructor(playerColor, mvEvtMsg){
        this.playerColor = playerColor;
        this.moveEventMessage = mvEvtMsg;
    }
}

class ErrorEventMessage {
    error: string;

    constructor(err){
        this.error = err;
    }
}

class WebSocketManager {
    private socket: WebSocket | null = null;

    connect(): void {
        this.socket = new WebSocket('ws://localhost:8080/ws');

        this.socket.addEventListener('open', () => {
            console.log('ws conn opened');
        });

        this.socket.addEventListener('message', (evt) => {
            const msg: EventMessage = JSON.parse(evt.data);
            routeEventMessage(msg);
        });

        this.socket.addEventListener('error', (err) => {
            console.error('resp error:', err)
        });

        this.socket.addEventListener('close', (c) => {
            console.log('ws conn closed', c);
            this.socket = null;
        });
    }

    send(evtMsg: EventMessage): void {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(evtMsg));
        } else {
            console.log('cannot send message websocket not open.')
        }
    }
}

function routeEventMessage(evtMsg: EventMessage) {
    if (evtMsg.type === undefined) {
        alert('no type field in the event');
    }

    switch(evtMsg.type) {
        default:
            console.log('resp:', evtMsg);
            break;
    }
}

async function getEventType(): Promise<string> {
    return new Promise((resolve) => {
        rl.question('event type: ', (resp) => {
            resolve(resp);
        });
    });
}

async function getEventPayload(): Promise<string> {
    return new Promise((resolve) => {
        rl.question('event payload:', (resp) => {
            resolve(resp);
        });
    });
}

const wsManager = new WebSocketManager();
wsManager.connect();

const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

async function main() {
    while (true) {
        var type = await getEventType();
        var payload = await getEventPayload();

        const evtMsg = new EventMessage(type, payload);
        console.log(JSON.stringify(evtMsg));
        wsManager.send(evtMsg);
    }

    rl.close();
}

main();
// join_match
// {"time_control":"1m"}

// make_move
// {"move":"e4"}