document.addEventListener('DOMContentLoaded', () => {

    console.log('Here we go!!!');

    const articles = document.getElementsByClassName('article');

    for (const article of articles) {
        // const img = gif[Math.floor(Math.random() * gif.length)]
        const img = article.getElementsByTagName('a')[0].href.split('/')[3];
        console.log(img);
        article.addEventListener('mouseenter', () => {
            document.body.style.backgroundImage = 'url(/image/' + img + '.gif)';
            document.body.style.backgroundSize = 'cover';
            document.body.style.backgroundPosition = 'center';
            document.body.style.backgroundAttachment = 'fixed';
        });
        article.addEventListener('mouseleave', () => {
            document.body.style.background = 'white';
        });
    }

    for (let i = 0; i < articles.length; i++) {
        const article = articles[i];
        const url = article.getElementsByTagName('a')[0].href.split('/')[3];
        article.addEventListener('mouseenter', () => {
            document.body.style.backgroundImage = 'url(/image/' + url + '.gif)';
            for (let j = 0; j < articles.length; j++) {
                if (i != j) {
                    articles[j].classList.toggle('hidden');
                }
            }
            article.classList.toggle('hover');
        });
        article.addEventListener('mouseleave', () => {
            document.body.style.backgroundImage = 'none';
            for (let j = 0; j < articles.length; j++) {
                articles[j].classList.toggle(i == j ? 'hover' : 'hidden');
            }
        });
    }

});

