var $grid = $('.grid').imagesLoaded(function () {
    $grid.masonry({
        itemSelector: '.grid-item',
        percentPosition: true,
        columnWidth: '.grid-sizer',
        gutter: 4
    });
});

$(document).ready(function () {
    $('.grid').Chocolat({
        displayAsALink: true,
        linkImages: true,
        imageSize: 'contain',
        enableZoom: true
    });
    $('a').hover(function () {
        $(this).attr('title', '');
    });

    var width = $(".grid-sizer").width();
    var highest = Number.NEGATIVE_INFINITY;
    var tmp;

    for (var image in images) {
        tmp = images[image].Ratio;
        if (tmp > highest) highest = tmp;
    }
    var minHeight = width / (Math.round(highest * 100) / 100);


    for (var key in images){
        var elem = $('div[name="' + images[key].Name + '"]');
        var height = Math.round(((elem.width() / images[key].Ratio) / minHeight));
        if (height > 1) {
            elem.height(height * minHeight + height * 4 - 4);
        } else {
            elem.height(height * minHeight);
        }
    }
    var myLazyLoad = new LazyLoad();
});

