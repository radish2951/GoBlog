document.addEventListener('DOMContentLoaded', () => {

    const articles = document.getElementsByClassName('article');

    for (let i = 0; i < articles.length; i++) {

        const article = articles[i];
        const url = article.getElementsByTagName('a')[0].href.split('/')[3];

        article.addEventListener('mouseenter', () => {
            document.body.style.backgroundImage = 'url(/image/' + url + '.gif)';
            for (let j = 0; j < articles.length; j++) {
                articles[j].classList.toggle(i == j ? 'hover' : 'hidden');
            }
        });

        article.addEventListener('mouseleave', () => {
            document.body.style.backgroundImage = 'none';
            for (let j = 0; j < articles.length; j++) {
                articles[j].classList.toggle(i == j ? 'hover' : 'hidden');
            }
        });

    }
});
