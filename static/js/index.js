var $grid = $('.grid').imagesLoaded( function() {
    // init Masonry after all images have loaded
    $grid.masonry({
        itemSelector: '.grid-item',
        percentPosition: true,
        columnWidth: '.grid-sizer',
        gutter: 4
    });
});

$(document).ready(function(){
    $('.grid').Chocolat({
        displayAsALink: true,
        linkImages : true,
        imageSize: 'contain',
        enableZoom: true
    });
    $('a').hover(function(){
        $(this).attr('title', '');
    });
});

$(window).imagesLoaded(function () {
    var viewportWidth = $(window).width();
    if (viewportWidth < 641) {
            $(".meta").addClass("grid-item--width2");
    }
});