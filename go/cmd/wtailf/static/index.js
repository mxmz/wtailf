
const result = document.getElementById("result");
var source = new EventSource("/events");
source.addEventListener("log", function (event) {

    const elem = document.createElement("div");
    elem.innerText = event.data;
    elem.classList.add("logline");
    result.insertAdjacentElement("afterbegin", elem);
    if (result.childNodes.length > 1000) {
        result.lastChild.remove();
    }
    
});
