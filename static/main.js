/**
 * Main JavaScript file for Spatially Invariant Vector Quantization
 */
var main = {

    /*
     * UI elements
     */
    canvasOriginal: null,
    divOriginal: null,
    divResult: null,
    formOptions: null,
    formUpload: null,
    inputVectorRadius: null,
    inputX: null,
    inputY: null,
    inputImageName: null,
    buttonStop: null,
    buttonSIVQ: null,
    buttonAdjustParameters: null,
    inputNewVectorName: null,
    buttonSaveNewVector: null,
    selectVector: null,
    inputAdvanced: null,
    divAdvancedOptions: null,
    divChooseBest: null,
    
    /*
     * Original image
     */
    imageOriginal: null,
    
    /*
     * Image uploaded
     */
    processUpload: function() {
        main.imageOriginal = null;
        
        // process response
        try {
            var response = $(this).contents().find("body").html();
            response = eval('(' + response + ')');
        } catch(e) {
            alert("Invalid response!\n"+ e);
            return;
        }
        
        // show error message
        if (response.Error) {
            main.showError(response.Message);
            return;
        }
        
        // prepare UI
        main.divOriginal.show();
        main.divResult.show().html('No results yet.');
        main.formOptions.show();
        main.inputImageName.val(response.Image);
        main.resizeResultView();

        // loading message
        main.canvasOriginal.width = 400;
        main.canvasOriginal.height = 50;
        var ctx = main.canvasOriginal.getContext("2d");
        ctx.fillStyle = "black";
        ctx.font="10pt Arial";
        ctx.textBaseline = "middle";
        ctx.clearRect(0,0,400,100);
        ctx.fillText("Loading image...", 10, 25);

        // load image
        var imageUrl = "/img/upload/"+ response.Image;
        main.imageOriginal = new Image();
        main.imageOriginal.onload = function() {
            var width = this.width;
            var height = this.height;
            
            // resize canvas
            main.canvasOriginal.width = width;
            main.canvasOriginal.height = height;
            
            // insert image to canvas
            ctx.drawImage(this, 0, 0);
        };
        main.imageOriginal.src = imageUrl;
    },
    
    resizeResultView: function() {
    	var viewHeight = $(window).height() - 20;
    	
    	// remove all other element heights
    	$("h1:first").nextUntil("#images").filter(":visible").add($("h1:first")).each(function() {
    		viewHeight -= $(this).outerHeight(true);
    	});
    	
    	if (viewHeight < 200) {
        	viewHeight = 200;
        }

        main.divOriginal.height(viewHeight);
        main.divResult.height(viewHeight);
    },

    showError: function(message) {
        var errorDiv = $(document.createElement("div")).addClass("error")
        					.insertBefore(main.formUpload).html(message);
        main.resizeResultView();

        setTimeout(function() {
        	errorDiv.remove();
        	errorDiv = null;
        	main.resizeResultView();
        }, 3000);
    },

    /*
     * Get coordinates from click on image 
     */
    coordinates: function(e) {
        if (main.imageOriginal == null) {
            return;
        }

        var offset = $(this).offset();
        var vectorX = e.pageX - offset.left;
        var vectorY = e.pageY - offset.top;

        main.inputX.val(vectorX);
        main.inputY.val(vectorY);
        
        main.drawVector();
    },

    /*
     * Draw vector on original image
     */
    drawVector: function(clear) {
        if (main.imageOriginal == null) {
            return;
        }
        
        // input
        var x = parseInt(main.inputX.val());
        var y = parseInt(main.inputY.val());
        var r = parseInt(main.inputVectorRadius.val());
        if ((x < 0 || y < 0 || r <= 0) && !clear) {
            return;
        }
        
        // draw vector
        var ctx = main.canvasOriginal.getContext("2d");
        ctx.clearRect(0, 0, main.imageOriginal.width, main.imageOriginal.height);
        ctx.drawImage(main.imageOriginal, 0, 0);
        
        if (clear != true) {
            ctx.beginPath();
            ctx.arc(x, y, r, 0, Math.PI*2, false); 
            ctx.closePath();
            ctx.stroke();
            
            main.selectVector.val("");
        }
    },
    
    /*
     * Vector input fields change event handler
     */
    vectorChanged: function(e) {
        if (this.value == "") {
            return;
        }
        main.drawVector();
    },
    
    init: function() {
        // UI elements
    	main.formUpload = $("#uploadForm");
        main.divOriginal = $("#original");
        main.divResult = $("#result");
        main.divChooseBest = $("#chooseBest");
        
        main.divOriginal.scroll(function(e){
            main.divResult.scrollLeft(main.divOriginal.scrollLeft());
            main.divResult.scrollTop(main.divOriginal.scrollTop());
        });
        
        main.divResult.scroll(function(e){
            main.divOriginal.scrollLeft(main.divResult.scrollLeft());
            main.divOriginal.scrollTop(main.divResult.scrollTop());
        });
        
        main.canvasOriginal = document.createElement("canvas");
        main.divOriginal.get(0).appendChild(main.canvasOriginal);
        main.formOptions = $("#optionsForm").submit(function(e) {
            e.preventDefault();
            process.process(false);
        });
        main.inputVectorRadius = $("#vectorRadius").keyup(main.vectorChanged);
        main.inputX = $("#vectorX").keyup(main.vectorChanged);
        main.inputY = $("#vectorY").keyup(main.vectorChanged);
        main.inputImageName = $("#imageName");
        main.buttonSIVQ = $("#sivq");
        main.buttonStop = $("#stop").click(function(e) {
            process.stop();
            return false;
        });
        main.buttonAdjustParameters = $("#adjustParameters").click(function(e) {
        	e.preventDefault();
        	e.stopPropagation();
        	process.process(true);
        });
        main.inputNewVectorName = $("#newVectorName");
        main.buttonSaveNewVector = $("#saveNewVector").click(function(e) {
            process.saveVector();
            return false;
        });
        main.selectVector = $("#vectorSelector").change(function(e) {
            main.drawVector(true);
        });
        
        $("#uploadResponse").load(main.processUpload);
        $("#original").delegate("canvas", "click", main.coordinates);
        
        $(window).resize(main.resizeResultView);
    }
};
