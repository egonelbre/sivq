/**
 * Process image
 */
var process = {

	/*
	 * Web Socket connection
	 */
	connection: null,

	adjustParameters: false,

	input: {},

	tryParameters: [],
	tryValues: [],

	currentParameter: "",
	currentValue: 0,
	
	parameters: {
		vectorRadius: [3,4,7,10],
		ringSizeInc: [1,2,3,5],
		vectorRings: [1,2,3],

		matchStride: [1,2,3,5],
		matchingOffset: [0,1,2],
		rotationStride: [0.001, Math.PI / 16, Math.PI / 8, Math.PI / 4],

		averageBias: [0.0, 0.3, 0.6, 1],
		gammaAdjust: [1.0, 2.0, 4.0, 8.0]
	},
	
	parameterName: {
		vectorRadius: "vector radius",
		ringSizeInc: "radius increment",
		vectorRings: "vector rings",
		
		matchStride: "matching value stride",
		matchingOffset: "matching offset",
		rotationStride: "rotation stride",
		
		averageBias: "average bias",
		gammaAdjust: "gamma adjust"
	},
	
	getInstructionText: function(parameter) {
		switch (parameter) {
		case "matchStride":
		case "matchingOffset":
		case "rotationStride":
			return "Optimizations. Choose higher if possible (faster).";
		case "averageBias":
		case "gammaAdjust":
			return "Post-processing";
		default:
			return "Choose best image";
		}
	},

	getInput: function() {
		var input = {
			image: main.inputImageName.val(),
			vectorName: $.trim(main.selectVector.val()),
			vecX: parseInt(main.inputX.val()),
			vecY: parseInt(main.inputY.val()),
			vectorRadius: parseInt(main.inputVectorRadius.val()),
			vectorRings: parseInt($("#vectorRings").val()),
			ringSizeInc: parseInt($("#ringSizeInc").val()),
			threshold: parseFloat($("#threshold").val()),
			rotationStride: parseFloat($("#rotationStride").val()),
			matchStride: parseInt($("#matchStride").val()),
			matchingOffset: parseInt($("#matchingOffset").val()),
			gammaAdjust: parseFloat($("#gammaAdjust").val()),
			averageBias: parseFloat($("#averageBias").val())
		};

		// remove NaNs
		for (i in input) {
			if (isNaN(input[i]) && i != "vectorName" && i != "image") {
				input[i] = -1;
			}
		}

		return input;
	},
	
	cloneArray: function(array) {
		var newArray = [];
		for (i in array) {
			newArray[i] = array[i];
		}
		return newArray;
	},

	/*
	 * Process image
	 */
	process: function(adjustParameters) {
		if (process.connection != null) {
			return;
		}

		var input = process.getInput();
		if (((input.vecX < 0 || input.vecY < 0 || input.vectorRadius <= 0 || input.vectorRings <= 0 || input.ringSizeInc < 0) && input.vectorName.length == 0)
				|| input.image.length == 0) {
			main.showError("Please fill in all fields.");
			return;
		}
		process.input = input;

		// prepare UI
		main.buttonSIVQ.attr("disabled", "disabled");
		main.buttonAdjustParameters.attr("disabled", "disabled");
		main.buttonStop.show();

		process.connection = new WebSocket("ws://localhost:8080/process");
		if (!process.connection) {
			main.showError("No connection!");
			return;
		}

		process.connection.onclose = function(e) {
	    	console.log("Connection closed.");
	    	process.connection = null;
	    };
	    process.connection.onmessage = process.serverMessage;

	    process.adjustParameters = adjustParameters;
	    process.connection.onopen = function() {
	    	if (!process.adjustParameters) {
		    	process.advancedProcess();
		    } else {
		    	process.adjustParametersProcess();
		    }
		};
	},
	
	advancedProcess: function() {
		main.divResult.html('<div class="loader"></div>');

		// send image for processing
		process.connection.send(JSON.stringify(process.input));
	},

	adjustParametersProcess: function() {
		// parameters to try
		process.tryParameters = [];
		for (i in process.parameters) {
			process.tryParameters.push(i);
		}
		console.log(process.tryParameters);

		process.nextParameter();
	},
	
	nextParameter: function() {
		process.currentParameter = process.tryParameters.shift();
		process.tryValues = process.cloneArray(process.parameters[process.currentParameter]);
		
		// some additional checks
		switch (process.currentParameter) {
		case "ringSizeInc":
			console.log("ringSizeInc, vectorRings = 3");
			process.input.vectorRings = 3;
			break;
		case "matchingOffset":
			if (process.input.matchStride != 3) {
				console.log("Skipping matchingOffset.");
				process.nextParameter();
				return;
			}
			break;
		}

		// loaders for images
		main.divResult.empty();
		var n = process.tryValues.length;
		var nSqrt = Math.sqrt(n);
		var rows = Math.round(nSqrt);
		var columns = (nSqrt == rows || Math.floor(nSqrt) != rows) ? rows : rows + 1;
		var imageWidth = (main.divResult.innerWidth() - 16) / columns - 1;
		var imageHeight = (main.divResult.innerHeight() - 16) / rows - 1;
		var i, divImage;
		for (i = 0; i < n; i++) {
			divImage = $(document.createElement("div")).addClass("variableSelect").width(imageWidth).height(imageHeight)
							.appendTo(main.divResult);
			$(document.createElement("div")).addClass("loader").width(0)
				.appendTo(divImage);
		}

		process.nextValue(null);
	},
	
	nextValue: function(data) {
		if (data != null) {
			main.divResult.find(".loader:first").parent().data("value", process.currentValue)
				.html('<img src="data:image/png;base64,'+ data +'" alt="" />');
		}

		if (process.tryValues.length == 0) {
			// show choose message
			main.divChooseBest.html(process.getInstructionText(process.currentParameter) +' <span class="bold">('
					+ process.parameterName[process.currentParameter] +')</span>:');
			main.divChooseBest.show().css("right", (main.divResult.width() - main.divChooseBest.width()) / 2);

			main.divResult.children(".variableSelect").css("cursor", "pointer")
				.click(function(e) { process.selectBest($(this)); });
			return;
		}

		// send new processing request
		var newInput = jQuery.extend({}, process.input);

		process.currentValue = process.tryValues.shift();
		newInput[process.currentParameter] = process.currentValue;
		console.log(process.currentParameter +" = "+ process.currentValue);

		// send image for processing
		process.connection.send(JSON.stringify(newInput));
	},
	
	selectBest: function(choice) {
		main.divChooseBest.hide();

		// update process input
		process.input[process.currentParameter] = choice.data("value");
		$("#"+ process.currentParameter).val(choice.data("value"));
		
		// all done
		if (process.tryParameters.length == 0) {
			var finalImage = choice.children("img:first");
			main.divResult.empty().append(finalImage);

			process.closeConnection();
			return;
		}

		process.nextParameter();
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

			if (process.adjustParameters) {
				process.nextValue(data);
				return;
			}

			main.divResult.html('<img src="data:image/png;base64,'+ data +'" alt="" />');
			process.closeConnection();
		} else {
			// loader status
			main.divResult.find(".loader:first").width(parseInt(parseFloat(data) * 100) + "%");
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

		$("div.variableSelect").unbind("click").css("cursor", "");
		main.divChooseBest.hide();
	},

	/*
	 * Close connection
	 */
	closeConnection: function() {
		console.log("Closing connection.");
		if (process.connection == null) {
			return;
		}

		try {
			process.connection.close();
			process.connection = null;
		} catch(e) {}

		main.buttonSIVQ.removeAttr("disabled");
		main.buttonAdjustParameters.removeAttr("disabled");
		main.buttonStop.hide();
	},
	
	/*
	 * Save vector into file
	 */
	saveVector: function() {
		var input = process.getInput();
		input.vectorName = $.trim(main.inputNewVectorName.val());
		
		if (input.vecX < 0 || input.vecY < 0 || input.vectorRadius <= 0 || input.vectorRings <= 0
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
		}, "json");
	}

};