const gameBoard = document.querySelector("#gameboard");
const playerDisplay = document.querySelector("#player");
const infoDisplay = document.querySelector("#info-display");
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

function checkIfValidMove(startId, targetId) {
    const piece = draggedElement.id;
    
    switch(true) {
        case piece.includes("pawn") :
            if ( piece.includes("light") && validLightPawnMove(startId, targetId, width) ) {
                // TODO light piece promotion if back rank
                return true;
            }
            
            if ( piece.includes("dark") && validDarkPawnMove(startId, targetId, width) ) {
                // TODO dark piece promotion if back rank
                return true;
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

var websocketDragDrop = function(wsManager) {
    return function dragDrop(e) {
        e.stopPropagation();

        const correctTurn = draggedElement.getAttribute('id').includes(playerTurn);
        if (!correctTurn) {
            temporaryMessage("not your turn buddy");
            return
        }
        
        const startId = Number(startPositionId);
        const targetId = Number(e.target.getAttribute("square-id") || e.target.parentNode.parentNode.getAttribute('square-id'));
        const valid = checkIfValidMove(startId, targetId);
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
        
        // TODO: convert move ot algebraic notation
        console.log("moved piece:", draggedElement.id);
        console.log("targeted piece if any:", e.target.parentNode.getAttribute("class"));
        console.log("start square:", squareIdToAlgebraicNotation(startId));
        console.log("target square:", squareIdToAlgebraicNotation(targetId));

        const evtMsg = new EventMessage("make_move", '{"move":"e4"}');
 
        wsManager.send(evtMsg);
        wsManager.interrupt()
            .then(() => {
                if (taken && containsOpponent) {
                    const square = e.target.parentNode.parentNode;
                    e.target.parentNode.remove();
                    square.append(draggedElement);
                } else {
                    e.target.append(draggedElement);
                }
                
                changePlayer();
            })
            .catch((error) => {
                temporaryMessage(JSON.stringify(error));
            });
    }
}