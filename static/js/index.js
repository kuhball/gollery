var $grid = $('.grid').imagesLoaded( function() {
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