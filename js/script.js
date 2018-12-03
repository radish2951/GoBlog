document.addEventListener('DOMContentLoaded', () => {

    const articles = document.getElementsByClassName('article');

    for (let i = 0; i < articles.length; i++) {

        const url = articles[i].getElementsByTagName('a')[0].href.split('/')[3];

        articles[i].addEventListener('mouseenter', () => {
            document.body.style.backgroundImage = 'url(/image/' + url + '.gif)';
            for (let j = 0; j < articles.length; j++) {
                articles[j].classList.toggle(i == j ? 'hover' : 'hidden');
            }
        });

        articles[i].addEventListener('mouseleave', () => {
            document.body.style.backgroundImage = 'none';
            for (let j = 0; j < articles.length; j++) {
                articles[j].classList.toggle(i == j ? 'hover' : 'hidden');
            }
        });

    }
});
