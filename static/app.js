const gameBoard = document.querySelector("#gameboard");
const playerDisplay = document.querySelector("#player");
const infoDisplay = document.querySelector("#info-display");
const width = 8;
let playerTurn = 'light';
playerDisplay.textContent = playerTurn;

// these can & probably should be dynamically created perspectives
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
            setBoard(piece, i);
        })
    }
}

function changePlayer() {

}

let startPositionId = null;
let draggedElement = null;

function dragStart(e) {
    draggedElement = e.target;
    startPositionId = e.target.parentNode.getAttribute('square-id');
}

function dragOver(e) {
    e.preventDefault();
}

function dragDrop(e) {
    e.stopPropagation();
    const taken = e.target.classList.contains('piece');

    // e.target.parentNode.append(draggedElement);
    // e.target.remove();
    changePlayer();
}

// createBoard()
createBoard("dark");

const allSquares = document.querySelectorAll("#gameboard .square");

allSquares.forEach( square => {
    square.addEventListener('dragstart', dragStart);
    square.addEventListener('dragover', dragOver);
    square.addEventListener('drop', dragDrop);
})
