const gameBoard = document.querySelector("#gameboard");
const playerDisplay = document.querySelector("#player");
const infoDisplay = document.querySelector("#info-display");

const width = 8;
const startPieces = [
    darkRook, darkKnight, darkBishop, darkQueen, darkKing, darkBishop, darkKnight, darkRook,
    darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn, darkPawn,
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    '', '', '', '', '', '', '', '',
    lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn, lightPawn,
    lightRook, lightKnight, lightBishop, lightQueen, lightKing, lightBishop, lightKnight, lightRook,
];

// createBoard()
function createBoardLightPerspective() {
    startPieces.forEach((piece, i) => {
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
    })
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

    //e.target.parentNode.append(draggedElement);
    e.target.append(draggedElement);
}

// createBoard()
createBoardLightPerspective();

const allSquares = document.querySelectorAll("#gameboard .square");

allSquares.forEach( square => {
    square.addEventListener('dragstart', dragStart);
    square.addEventListener('dragover', dragOver);
    square.addEventListener('drop', dragDrop);
})
