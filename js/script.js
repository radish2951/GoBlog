window.addEventListener('load', () => {

    const articles = document.getElementsByClassName('article');
    const articleArray = Array.from(articles);

    let x = 0;
    let y = 0;

    for (let i = 0; i < articles.length; i++) {

        const url = articles[i].getElementsByTagName('a')[0].href.split('/')[3];

        articles[i].addEventListener('mouseenter', () => {
            document.body.style.backgroundImage = 'url(/image/' + url + '.gif)';
            for (let j = 0; j < articles.length; j++) {
                articles[j].classList.add(i == j ? 'hover' : 'hidden');
            }
        });

        articles[i].addEventListener('mouseleave', () => {
            document.body.style.backgroundImage = 'none';
            for (let j = 0; j < articles.length; j++) {
                articles[j].classList.remove('hover', 'hidden');
            }
        });

    }

    document.addEventListener('mousemove', (e) => {

        x = e.clientX;
        y = e.clientY;

    });

    window.addEventListener('scroll', () => {

        const element = document.elementFromPoint(x, y);
        const article = element.parentElement.parentElement || document.body;
        const index = articleArray.indexOf(article);

        if (index + 1) {
            const url = article.getElementsByTagName('a')[0].href.split('/')[3];
            document.body.style.backgroundImage = 'url(/image/' + url + '.gif)';
            for (let i = 0; i < articles.length; i++) {
                articles[i].classList.add(i == index ? 'hover' : 'hidden');
            }
        } else {
            document.body.style.backgroundImage = 'none';
            for (let i = 0; i < articles.length; i++) {
                articles[i].classList.remove('hover', 'hidden');
            }
        }

    });

});
