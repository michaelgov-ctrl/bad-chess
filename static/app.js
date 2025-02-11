const gameBoard = document.querySelector("#gameboard");
const playerDisplay = document.querySelector("#player");
const infoDisplay = document.querySelector("#info-display");
const width = 8;

// these can & probably should be dynamically created perspectives, maybe...?
const startPiecesLightPerspective = [
    darkRook, darkKnight, darkBishop, darkQueen, darkKing, darkBishop, darkKnight, darkRook,
    darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn,
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn,
    lightRook, lightKnight, lightBishop, lightQueen, lightKing, lightBishop, lightKnight, lightRook,
];

const startPiecesDarkPerspective = [
    lightRook, lightKnight, lightBishop, lightKing, lightQueen, lightBishop, lightKnight, lightRook,
    lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn,
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn,
    darkRook, darkKnight, darkBishop, darkKing, darkQueen, darkBishop, darkKnight, darkRook,
];

let playerTurn = 'light';
playerDisplay.textContent = playerTurn;
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

function reverseIds() {
    const allSquares = document.querySelectorAll(".square");
    allSquares.forEach((square, i) => {
        square.setAttribute('square-id', (width * width - 1) - i);
    })
}

function revertIds() {
    const allSquares = document.querySelectorAll(".square");
    allSquares.forEach((square, i) => {
        square.setAttribute('square-id', i);
    })
}

function changePlayer() {
    if (playerTurn === "dark") {
        //revertIds(); // only necessary for shared play
        playerTurn = "light";
        playerDisplay.textContent = "light";
    } else {
        //reverseIds(); // only necessary for shared play
        playerTurn = "dark";
        playerDisplay.textContent = "dark";
    }
}

function checkIfValidMove(target) {
    const targetId = Number(target.getAttribute("square-id") || target.parentNode.parentNode.getAttribute('square-id'));
    const startId = Number(startPositionId);
    const piece = draggedElement.id;
    
    switch(true) {
        case piece.includes("pawn") :
            if (piece.includes("light")) {
                const starterRow = [48,49,50,51,52,53,54,55];
                if (
                    (starterRow.includes(startId) && startId - (width * 2) === targetId) ||
                    startId - width === targetId ||
                    (startId - width - 1 === targetId && document.querySelectorAll(`[square-id="${startId - width - 1}"]`)[0]?.firstChild?.getAttribute("class") === "piece") ||
                    (startId - width + 1 === targetId && document.querySelectorAll(`[square-id="${startId - width + 1}"]`)[0]?.firstChild?.getAttribute("class") === "piece")
                ) {
                    return true
                }
            } else {
                const starterRow = [8,9,10,11,12,13,14,15];
                if (
                    (starterRow.includes(startId) && startId + (width * 2) === targetId) ||
                    startId + width === targetId ||
                    (startId + width - 1 === targetId && document.querySelectorAll(`[square-id="${startId + width - 1}"]`)[0]?.firstChild?.getAttribute("class") === "piece") ||
                    (startId + width + 1 === targetId && document.querySelectorAll(`[square-id="${startId + width + 1}"]`)[0]?.firstChild?.getAttribute("class") === "piece")
                ) {
                    return true
                }
            }
            break;
        case piece.includes("knight") :
            if (
                startId + (width * 2) - 1 === targetId ||
                startId + (width * 2) + 1 === targetId ||
                startId + width - 2 === targetId ||
                startId + width + 2 === targetId ||
                startId - (width * 2) - 1 === targetId ||
                startId - (width * 2) + 1 === targetId ||
                startId - width - 2 === targetId ||
                startId - width + 2 === targetId
            ) {
                return true
            }
            break;
        case piece.includes("bishop") :
            
            return true
        case piece.includes("rook") :
            return true
        case piece.includes("queen") :
            return true
        case piece.includes("king") :
            return true
    }

    console.log("out of piece validation switch statement")
    return false
}

function temporaryMessage(msg) {
    infoDisplay.textContent = msg;
    setTimeout(() => infoDisplay.textContent = "", 2000)
}

function dragStart(e) {
    draggedElement = e.target;
    startPositionId = e.target.parentNode.getAttribute('square-id');
}

function dragOver(e) {
    e.preventDefault();
}

function dragDrop(e) {
    e.stopPropagation();

    const correctTurn = draggedElement.getAttribute('id').includes(playerTurn);
    if (!correctTurn) {
        temporaryMessage("not your turn buddy");
        return
    }
    
    const valid = checkIfValidMove(e.target);
    if (!valid) {
        //console.log("taken", taken, "opp", opponent, "takenBy", takenByOpponent, "v", valid);
        temporaryMessage("invalid move");
        return
    }

    const taken = e.target.parentNode.getAttribute("class")?.includes("piece");//?.contains('piece');
    const opponent = playerTurn === "light" ? "dark" : "light";
    const containsOpponent = e.target.parentNode.getAttribute("id").includes(opponent);
    if (taken && !containsOpponent) {
        temporaryMessage("invalid move");
        return
    }    
    
    // send move to server for validation here, if error return

    if (taken && containsOpponent) {
        const square = e.target.parentNode.parentNode;
        e.target.parentNode.remove();
        square.append(draggedElement);
    } else {
        e.target.append(draggedElement);
    }
    
    changePlayer();
}



// start
createBoard("dark");

const allSquares = document.querySelectorAll(".square");

allSquares.forEach( square => {
    square.addEventListener('dragstart', dragStart);
    square.addEventListener('dragover', dragOver);
    square.addEventListener('drop', dragDrop);
})
