$jq("#eventSave").click(function() {
    var momaFormArray = CheckFormToArray($jq("#momaform"));
    var formData = JSON.stringify(momaFormArray);
    $jq.ajax({
        url: "/events",
        type: 'POST',
        contentType: 'application/json; charset=UTF-8',
        data: formData,
        success: function() {
            document.location='/';
        }
    });
});

function deleteEvent(eventId){

    $jq.ajax({
        url: "/events/"+eventId,
        type: 'DELETE',
        success: function() {
            document.location='/';
        }
    });
};

/* Check form content */
function CheckFormToArray(form){
    var listOfAvailableFields = ["Email", "Who", "What", "When", "Lat", "Lng", "Pic"]
    var array = jQuery(form).serializeArray();
    var checkedForm = {};
    jQuery.each(array, function() {
        if(jQuery.inArray(this.name, listOfAvailableFields) >= 0){
            switch(this.name){
                // Special case for time
                case 'When':
                    if(this.value == ''){
                        this.value = Date.now();
                    }
                    convertDate = new Date(this.value);
                    this.value = ISODateString(convertDate);
                    checkedForm['When'] = this.value;
                    break;
                // Special case for pics
                case 'Pic':
                    if(!checkedForm.hasOwnProperty('Pic')){
                            checkedForm['Pic'] = [];
                    }
                    checkedForm['Pic'].push(this.value);
                    break;
                // Special case for where
                case 'Lat':
                    if(!checkedForm.hasOwnProperty('Where')){
                            checkedForm['Where'] = {"Lng":0, "Lat":0};
                    }
                    checkedForm['Where'].Lat = parseFloat(this.value);
                    break;
                case 'Lng':
                    if(!checkedForm.hasOwnProperty('Where')){
                            checkedForm['Where'] = {"Lng":0, "Lat":0};
                    }
                    checkedForm['Where'].Lng = parseFloat(this.value);
                    break;
                default:
                    checkedForm[this.name] = this.value;
            }
        }
    });

    return checkedForm
}

/* use a function for the exact format desired... */
function ISODateString(d){
 function pad(n){return n<10 ? '0'+n : n}
 return d.getUTCFullYear()+'-'
      + pad(d.getUTCMonth()+1)+'-'
      + pad(d.getUTCDate())+'T'
      + pad(d.getUTCHours())+':'
      + pad(d.getUTCMinutes())+':'
      + pad(d.getUTCSeconds())+'Z'
}

var current_fs, next_fs, previous_fs; // fieldsets
var left, opacity, scale; // fieldset properties which we will animate
var animating; // flag to prevent quick multi-click glitches
var mapDisplayed = 0;

$jq(".next").click(function(){
    if(animating) return false;
    animating = true;

    current_fs = $jq(this).parent();
    next_fs = $jq(this).parent().next();

    // activate next step on progressbar using the index of next_fs
    $jq("#progressbar li").eq($jq("fieldset").index(next_fs)).addClass("active");

    //show the next fieldset
    next_fs.show();
    //hide the current fieldset with style
    current_fs.animate({opacity: 0}, {
        step: function(now, mx) {
            //as the opacity of current_fs reduces to 0 - stored in "now"
            //1. scale current_fs down to 80%
            scale = 1 - (1 - now) * 0.2;
            //2. bring next_fs from the right(50%)
            left = (now * 50)+"%";
            //3. increase opacity of next_fs to 1 as it moves in
            opacity = 1 - now;
            current_fs.css({'transform': 'scale('+scale+')'});
            next_fs.css({'left': left, 'opacity': opacity});
        },
        duration: 800,
        complete: function(){
            current_fs.hide();
            animating = false;
            if($jq("#mapbox").css('display') == 'block' && mapDisplayed == 0){initMap()};
        },
        //this comes from the custom easing plugin
        easing: 'easeInOutBack'
    });
});

$jq(".previous").click(function(){
    if(animating) return false;
    animating = true;

    current_fs = $jq(this).parent();
    previous_fs = $jq(this).parent().prev();

    //de-activate current step on progressbar
    $jq("#progressbar li").eq($jq("fieldset").index(current_fs)).removeClass("active");

    //show the previous fieldset
    previous_fs.show();
    //hide the current fieldset with style
    current_fs.animate({opacity: 0}, {
        step: function(now, mx) {
            //as the opacity of current_fs reduces to 0 - stored in "now"
            //1. scale previous_fs from 80% to 100%
            scale = 0.8 + (1 - now) * 0.2;
            //2. take current_fs to the right(50%) - from 0%
            left = ((1-now) * 50)+"%";
            //3. increase opacity of previous_fs to 1 as it moves in
            opacity = 1 - now;
            current_fs.css({'left': left});
            previous_fs.css({'transform': 'scale('+scale+')', 'opacity': opacity});
        },
        duration: 800,
        complete: function(){
            current_fs.hide();
            animating = false;

        },
        //this comes from the custom easing plugin
        easing: 'easeInOutBack'
    });
});

$jq("#addEvent").click(function(){
    document.location='#eventformmodal';
    return false;
});
