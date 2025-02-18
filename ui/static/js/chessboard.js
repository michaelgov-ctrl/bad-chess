const matchInfoDisplay = document.querySelector("#match-info-display");
const gameBoard = document.querySelector("#gameboard");
const playerDisplay = document.querySelector("#player");
const infoDisplay = document.querySelector("#info-display");
const playerClock = document.querySelector("#player_clock");
const opponentClock = document.querySelector("#opponent_clock");
const width = 8;

// these can & probably should be dynamically created perspectives, maybe...?
const startPiecesLightPerspective = [
    dark_rook, dark_knight, dark_bishop, dark_queen, dark_king, dark_bishop, dark_knight, dark_rook,
    dark_pawn, dark_pawn, dark_pawn, dark_pawn, dark_pawn, dark_pawn, dark_pawn, dark_pawn,
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    light_pawn, light_pawn, light_pawn, light_pawn, light_pawn, light_pawn, light_pawn, light_pawn,
    light_rook, light_knight, light_bishop, light_queen, light_king, light_bishop, light_knight, light_rook,
];

const startPiecesDarkPerspective = [
    light_rook, light_knight, light_bishop, light_king, light_queen, light_bishop, light_knight, light_rook,
    light_pawn, light_pawn, light_pawn, light_pawn, light_pawn, light_pawn, light_pawn, light_pawn,
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    dark_pawn, dark_pawn, dark_pawn, dark_pawn, dark_pawn, dark_pawn, dark_pawn, dark_pawn,
    dark_rook, dark_knight, dark_bishop, dark_king, dark_queen, dark_bishop, dark_knight, dark_rook,
];

let playerTurn = 'light';
playerDisplay.textContent = playerTurn;
let playerPieces = null;
let startPositionId = null;
let draggedElement = null;

function setBoard(piece, i) {
    const square = document.createElement('div');
    square.classList.add('square');
    square.innerHTML = piece;
    square.firstChild?.setAttribute('draggable', true)
    square.setAttribute('square-id', i);

    const row = Math.floor( (63 - i) / 8 ) + 1
    if ( row % 2 === 0 ) {
        square.classList.add(i % 2 === 0 ? "light" : "dark" );
    } else {
        square.classList.add(i % 2 === 0 ? "dark" : "light" );
    }

    gameBoard.append(square);
}

function createBoard(perspective) {
    playerPieces = perspective;

    if ( perspective === "light" ) {
        startPiecesLightPerspective.forEach((piece, i) => {
            setBoard(piece, i);
        })
    } else {
        /// if any perspective other than dark gets passed in it will be Dark Perspective as well
        startPiecesDarkPerspective.forEach((piece, i) => {
            setBoard(piece, 63-i);
        })
    }
}

function changePlayer() {
    if (playerTurn === "dark") {
        playerTurn = "light";
        playerDisplay.textContent = "light";
    } else {
        playerTurn = "dark";
        playerDisplay.textContent = "dark";
    }
}

function squareIdToAlgebraicNotation(squareId) {
    const file = String.fromCharCode(97 + (squareId % 8));
    const rank = Math.floor((63-squareId) / 8) + 1;
    return file + rank
}

function castleToAlgebraicNotation(startId, targetId, playerColor) {
    let defaultStart = 60;
    let defaultKingSide = 62;
    let defaultQueenSide = 58;
    if ( playerColor === "dark" ) {
        defaultStart = 4;
        defaultKingSide = 6;
        defaultQueenSide = 2;
    }

    if ( startId === defaultStart ) {
        switch (targetId) {
            case defaultKingSide:
                return "O-O";
            case defaultQueenSide:
                return "O-O-O";
        }
    }

    return;
}

function checkIfValidMove(startId, targetId, playerColor) {
    const piece = draggedElement.id;
    
    switch(true) {
        case piece.includes("pawn") :
            switch (playerColor) {
                case "light":
                    if ( validLightPawnMove(startId, targetId, width) ) {
                        return true;
                    }
                case "dark":
                    if ( validDarkPawnMove(startId, targetId, width) ) {
                        return true;
                    }
            }
            break;
        case piece.includes("knight") :
            if ( validKnightMove(startId, targetId, width) ) {
                return true;
            }
            break;
        case piece.includes("bishop") :
            if ( validBishopMove(startId, targetId, width) ) {
                return true;
            }
            break;
        case piece.includes("rook") :
            if ( validRookMove(startId, targetId, width) ) {
                return true;
            }
            break;
        case piece.includes("queen") :
            if ( validBishopMove(startId, targetId, width) || validRookMove(startId, targetId, width) ) {
                return true;
            }
            break;
        case piece.includes("king") :
            if ( validKingMove(startId, targetId, width) ) {
                return true;
            }
            break;
    }

    return false;
}

function pieceToLetter(p) {
    switch (p) {
        case "king":
            return "K";
        case "queen":
            return "Q";
        case "rook":
            return "R";
        case "bishop":
            return "B";
        case "knight":
            return "N";
        default:
            return "";
    }
}

function temporaryMessage(msg) {
    infoDisplay.textContent = msg;
    setTimeout(() => infoDisplay.textContent = "", 2000)
}

function showPromotionWindow() {
    return new Promise((resolve) => {
        document.getElementById("promotion-window").style.display = "block";

        document.getElementById("promote-to-queen").onclick = () => resolveChoice("Q", resolve);
        document.getElementById("promote-to-rook").onclick = () => resolveChoice("R", resolve);
        document.getElementById("promote-to-bishop").onclick = () => resolveChoice("B", resolve);
        document.getElementById("promote-to-knight").onclick = () => resolveChoice("N", resolve);
    });
}

function resolveChoice(choice, resolve) {
    hidePromotionWindow();
    resolve(choice);
}

function hidePromotionWindow() {
    document.getElementById("promotion-window").style.display = "none";
}

function handleChoice(choice) {
    hidePromotionWindow();
}

function dragStart(e) {
    draggedElement = e.target;
    startPositionId = e.target.parentNode.getAttribute('square-id');
}

function dragOver(e) {
    e.preventDefault();
}

var websocketDragDrop = function(gameManager) {
    return async function dragDrop(e) {
        e.stopPropagation();

        if ( !draggedElement.getAttribute('id').includes(playerTurn) ) {
            temporaryMessage("not your turn buddy");
            return
        }
        
        const startId = Number(startPositionId);
        const targetId = Number(e.target.getAttribute("square-id") || e.target.parentNode.parentNode.getAttribute('square-id'));
        if ( !checkIfValidMove(startId, targetId, playerTurn) ) {
            temporaryMessage("invalid move");
            return
        }

        let taken = e.target.parentNode.getAttribute("class")?.includes("piece");
        //const opponent = playerTurn === "light" ? "dark" : "light";
        const opponent = playerPieces === "light" ? "dark" : "light";
        const containsOpponent = e.target.parentNode.getAttribute("id").includes(opponent);
        if ( taken && !containsOpponent ) {
            temporaryMessage("invalid move");
            return
        }    
        
        // generate string for alebraic notation of the move to send to server
        let movedPiece = draggedElement.id.substring(draggedElement.id.indexOf("_") + 1);
        let pieceChar = pieceToLetter(movedPiece);
        let targetSquare = squareIdToAlgebraicNotation(targetId);

        let algMove = pieceChar + (taken ? "x" : "") + targetSquare;
        switch (pieceChar) {
            case "":
                if ( validEnPassant(startId, targetId, width) || taken ) {
                    taken = true;
                    pieceChar = squareIdToAlgebraicNotation(startPositionId).charAt(0);
                }
                algMove = pieceChar + (taken ? "x" : "") + targetSquare;
                
                if ( validPromotion(targetId, playerTurn) ) {
                    let popUpValue = await showPromotionWindow();
                    algMove += `=${popUpValue}`;
                }
                break;
            case "K":
                let castleAlg = castleToAlgebraicNotation(startId, targetId, playerTurn);
                algMove = castleAlg ? castleAlg : algMove;
                break;
        }

        let evtMsg = new EventMessage("make_move", `{"move":"${algMove}"}`);

        gameManager.send(evtMsg);
        gameManager.interrupt()
            .catch((error) => {
                temporaryMessage(JSON.stringify(error));
            });
    }
}
