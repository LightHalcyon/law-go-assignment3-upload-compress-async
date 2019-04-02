// const url = process.env.URL;
const url = 'http://localhost:20606/compress';

const form = document.querySelector('form');

form.addEventListener('submit', e => {
    e.preventDefault();

    const files = document.querySelector('[type=file]').files;
    const formData = new FormData();
    formData.append('file', files);
    fetch(url, {
        method: 'POST',
        headers: {
            "X-ROUTING-KEY": "asdf",
        },
        body: formData,
    }).then(response => {
        console.log(response);
    });
});