(function() {
    var Shorten = {};

    Shorten.submit = function(event) {
    
        url = $("input#longurl").val();

        var endpoint = '/enc?url=' + url;

        if (event){
            event.preventDefault();
        }

        $.ajax({
            url: endpoint,
            type: 'GET',
            success: function (resp) {
                callback(resp);
            },
            error: function (jqXHR, textStatus) {
                callback(jqXHR);
            }
        });

        callback = function(data) {
            if(data) {
                if(data.status == 'success') {
                    var result = data.data;
                    if( result === null) {
                        return;
                    }
                    $("#shorten-message").attr("class", "success");
                    $('#shorten-message').html('<a href="' + result + '">' + result + '</a>');
                } else {
                    $("#shorten-message").attr("class", "error");
                    $('#shorten-message').html('<p>' + data.responseJSON.error.message + '</p>');
                }
            }
            else {
                return false;
            }
        }
    }

    Shorten.init = function(event) {
        $(document).on("submit", ".form-shorten", Shorten.submit);
    };

    Shorten.init();

})();
