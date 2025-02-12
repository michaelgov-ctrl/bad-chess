// start
createBoard("light");

const allSquares = document.querySelectorAll(".square");

allSquares.forEach( square => {
    square.addEventListener('dragstart', dragStart);
    square.addEventListener('dragover', dragOver);
    square.addEventListener('drop', dragDrop);
})