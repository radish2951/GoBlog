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

window.addEventListener('load', () => {

    let x = 0;
    let y = 0;
    let lastArticle = '';

    document.addEventListener('mousemove', (e) => {
        x = e.clientX;
        y = e.clientY;
    });

    window.addEventListener('scroll', () => {
        /*
        const element = document.elementFromPoint(x, y);
        const article = element.parentElement.parentElement;
        if (article.classList.contains('article')) {
            article.classList.add('hover');
            lastArticle = article;
        } else {
            lastArticle.classList.remove('hover');
        }
        */
        console.log(window.scrollY);
    });
});
