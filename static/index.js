// const url = process.env.URL;
// dec2hex :: Integer -> String
function dec2hex (dec) {
    return ('0' + dec.toString(16)).substr(-2)
}
  
// generateId :: Integer -> String
function generateId (len) {
    var arr = new Uint8Array((len || 40) / 2)
    window.crypto.getRandomValues(arr)
    return Array.from(arr, dec2hex).join('')
}

const host = window.location.href.replace("http://", '');

const url = 'http://localhost:20606/compress';

const id = generateId()

const form = document.querySelector('form');

form.addEventListener('submit', e => {
    e.preventDefault();

    const files = document.querySelector('[type=file]').files[0];
    const formData = new FormData();
    formData.append('file', files);
    fetch(url, {
        method: 'POST',
        headers: {
            "X-ROUTING-KEY": id,
        },
        body: formData,
    }).then(response => {
        console.log(response);
    });
});