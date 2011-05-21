/**
 * Main JavaScript file for Spatially Invariant Vector Quantization
 */
var main = {

	/*
	 * Image uploaded
	 */
	processUpload: function() {
		// process response
		try {
			var response = $(this).contents().find("body").html();
			response = eval('(' + response + ')');
		} catch(e) {
			alert("Invalid response!\n"+ e);
			return;
		}

		// show error message, when there is error
		var errorDiv = $("#uploadForm").prev(".error").hide();
		if (response.Error) {
			if (errorDiv.size() == 0) {
				errorDiv = $(document.createElement("div")).addClass("error")
								.insertBefore($("#uploadForm"));
			}
			errorDiv.show().html(response.Message);
			return;
		}

		$("#original").show().html('Loading...').css({width: "auto", height: "auto"});
		$("#result").show().html('No results yet.').css({width: "auto", height: "auto"});
		$("#optionsForm").show().find("input:text").val("");
		$("#imageName").val(response.Image);

		var imageUrl = "/img/"+ response.Image;
		var loadImage = new Image();
		loadImage.onload = function() {
			var width = loadImage.width;
			var height = loadImage.height;

			console.log("Image ("+ width +" x "+ height +"; "+ imageUrl +")");

			$("#original").show().html('<img src="'+ imageUrl +'" alt="" />')
				.css({width: width, height: height});
		};
		loadImage.src = imageUrl;
	},
	
	/*
	 * Get resulting image
	 */
	calcResult: function(form) {
		$("#result").show().html('Loading...').css({width: "auto", height: "auto"});
		
		var imageUrl = "/result/"+ $("#imageName").val() +"?x="+ parseInt($("#vectorX").val())
							+"&y="+ parseInt($("#vectorY").val()) +"&r="+ parseInt($("#vectorRadius").val());
		var loadImage = new Image();
		loadImage.onload = function() {
			var width = loadImage.width;
			var height = loadImage.height;

			console.log("Result image ("+ width +" x "+ height +"; "+ imageUrl +")");

			$("#result").show().html('<img src="'+ imageUrl +'" alt="" />')
				.css({width: width, height: height});
		};
		loadImage.src = imageUrl;
	},

	/*
	 * Get coordinates from click on image 
	 */
	coordinates: function(e) {
		var offset = $(this).offset();
		var vectorX = e.pageX - offset.left;
		var vectorY = e.pageY - offset.top;
		console.log("Click at "+ vectorX +" "+ vectorY +"");
		$("#vectorX").val(vectorX);
		$("#vectorY").val(vectorY);
	},

	init: function() {
		$("#uploadResponse").load(main.processUpload);
		$("#original").delegate("img", "click", main.coordinates);
		$("#optionsForm").submit(function(e) {
			e.preventDefault();
			main.calcResult($(this));
		});
	}
};