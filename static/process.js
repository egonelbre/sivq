/**
 * Process image
 */
var process = {

	/*
	 * Web Socket connection
	 */
	connection: null,
	
	advanced: false,

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
		}

		// prepare UI
		main.buttonSIVQ.attr("disabled", "disabled");
		main.buttonStop.show();

		process.connection = new WebSocket("ws://localhost:8080/process");
		if (!process.connection) {
			alert("No connection!");
			return;
		}

		process.connection.onclose = function(e) {
	    	console.log("Connection closed.");
	    	process.connection = null;
	    };
	    process.connection.onmessage = process.serverMessage;

	    process.advanced = main.inputAdvanced.is(":checked");
	    if (process.advanced) {
	    	process.advancedProcess(input);
	    } else {
	    	process.multiStepProcess(input);
	    }
	},
	
	advancedProcess: function(input) {
		main.divResult.html('<div class="loader"></div>');

		// send image for processing
	    process.connection.onopen = function() {
			process.connection.send(JSON.stringify(input));
		};
	},

	multiStepProcess: function(input) {
		main.divResult.empty();

		var n = 4;
		var imageHeight = main.divResult.innerHeight() / 2;
		var i, divImage;
		for (i = 0; i < 4; i++) {
			divImage = $(document.createElement("div")).addClass("variableSelect").height(imageHeight)
							.appendTo(main.divResult);
			$(document.createElement("div")).addClass("loader").width(i +"%")
				.appendTo(divImage);
		}
		
		process.closeConnection();
	},

	/*
	 * Message from server
	 */
	serverMessage: function(e) {
		var data = e.data;
		if (data.substr(0, 1) == "{") {
			// error message
			data = JSON.parse(data);
			main.showError(data.Message);
			main.divResult.html(data.Message);
		} else if (data.length > 6) {
			// image ready
			main.divResult.html('<img src="data:image/png;base64,'+ data +'" alt="" />');
			process.closeConnection();
		} else {
			// loader status
			if (process.advanced) {
				main.divResult.children().first().width(parseInt(parseFloat(data) * 100) + "%");
			} else {
				// TODO: multi-step loader
			}
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
		process.closeConnection();
	},

	/*
	 * Close connection
	 */
	closeConnection: function() {
		if (process.connection == null) {
			return;
		}

		try {
			process.connection.close();
			process.connection = null;
		} catch(e) {}

		main.buttonSIVQ.removeAttr("disabled");
		main.buttonStop.hide();
	},
	
	/*
	 * Save vector into file
	 */
	saveVector: function() {
		var input = process.getInput();
		input.vectorName = $.trim(main.inputNewVectorName.val());
		
		if (input.vecX < 0 || input.vecY < 0 || input.radius <= 0 || input.vectorRings <= 0
				|| input.image.length == 0 || input.vectorName.length == 0) {
			main.showError("Could not save vector. Please make sure you have correctly selected vector.");
			return;
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