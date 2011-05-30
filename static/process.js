/**
 * Process image
 */
var process = {

	/*
	 * Web Socket connection
	 */
	connection: null,
	
	/*
	 * Loader bar
	 */
	divLoader: null,

	/*
	 * Process image
	 */
	process: function() {
		if (process.connection != null) {
			return;
		}

		var input = process.getInput();
		if (((input.vecX < 0 || input.vecY < 0 || input.radius <= 0 || input.vectorRings <= 0) && input.vectorName.length == 0)
				|| input.image.length == 0) {
			main.showError("Please fill in all fields.");
			return;
		} else {
			main.hideError();
		}
		console.log(input);

		// UI
		main.buttonSIVQ.attr("disabled", "disabled");
		main.buttonStop.show();
		main.divResult.show();

		process.connection = new WebSocket("ws://localhost:8080/process");
		if (!process.connection) {
			alert("No connection!");
			return;
		}

		process.connection.onclose = function(e) {
	    	console.log("Connection closed.");
	    	process.connection = null;
	    	
	    	main.buttonSIVQ.removeAttr("disabled");
			main.buttonStop.hide();
	    };

	    process.connection.onmessage = process.serverMessage;
		
		// send image for processing
	    process.connection.onopen = function() {
			process.connection.send(JSON.stringify(input));
		}

	    // loader
	    process.divLoader = $(document.createElement("div")).addClass("loader");
	    main.divResult.show().empty()
	    	.append(process.divLoader);
	},

	/*
	 * Message from server
	 */
	serverMessage: function(e) {
		var data = e.data;
		console.log(data);
		if (data.substr(0, 1) == "{") {
			data = JSON.parse(data);
			process.divLoader.width("100%");
			
			// load and show result image
			var resultImage = new Image();
			resultImage.onload = function() {
				main.divResult.html('<img src="'+ resultImage.src +'" alt="" />')
			};
			resultImage.src = "/img/result/"+ data.Image;
		} else {
			process.divLoader.width(parseInt(parseFloat(data) * 100) + "%");
		}
	},

	/*
	 * Stop processing image
	 */
	stop: function() {
		if (process.connection == null) {
			return;
		}
		process.connection.send("stop");
	},
	
	saveVector: function() {
		var input = process.getInput();
		input.vectorName = $.trim(main.inputNewVectorName.val());
		
		if (input.vecX < 0 || input.vecY < 0 || input.radius <= 0 || input.vectorRings <= 0
				|| input.image.length == 0 || input.vectorName.length == 0) {
			main.showError("Could not save vector. Please make sure you have correctly selected vector.");
			return;
		} else {
			main.hideError();
		}

		$.post("/saveVector", input, function(response) {
			if (response.Error) {
				main.showError(response.Message);
				return;
			}
			main.selectVector.append('<option value="'+ input.vectorName +'">'+ input.vectorName +'</option>');
			main.selectVector.val(input.vectorName);
			main.inputNewVectorName.val("");
			console.log(response);
		}, "json");
	},

	/*
	 * Create object of input values
	 */
	getInput: function() {
		var input = {
			image: main.inputImageName.val(),
			vectorName: $.trim(main.selectVector.val()),
			vecX: parseInt(main.inputX.val()),
			vecY: parseInt(main.inputY.val()),
			radius: parseInt(main.inputRadius.val()),
			vectorRings: parseInt($("#vectorRings").val()),
			ringSizeInc: parseInt($("#ringSizeInc").val()),
			threshold: parseFloat($("#threshold").val()),
			rotationStride: parseFloat($("#rotationStride").val()),
			matchStride: parseInt($("#matchStride").val()),
			matchingOffset: parseInt($("#matchingOffset").val()),
			gammaAdjust: parseFloat($("#gammaAdjust").val())
		};

		// remove NaNs
		for (i in input) {
			if (isNaN(input[i]) && i != "vectorName" && i != "image") {
				input[i] = -1;
			}
		}

		return input;
	}

};