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
	
	process: function() {
		if (process.connection != null) {
			return;
		}

		// input
		var input = {
				vecX: parseInt(main.inputX.val()),
				vecY: parseInt(main.inputY.val()),
				radius: parseInt(main.inputRadius.val()),
				image: main.inputImageName.val(),
				vectorRings: parseInt($("#vectorRings").val()),
				ringSizeInc: parseInt($("#ringSizeInc").val()),
				threshold: parseFloat($("#threshold").val()),
				rotationStride: parseFloat($("#rotationStride").val()),
				matchStride: parseInt($("#matchStride").val()),
				matchingOffset: parseInt($("#matchingOffset").val()),
				gammaAdjust: parseFloat($("#gammaAdjust").val())
			};
		if (isNaN(input.vecX) || input.vecX < 0 || isNaN(input.vecY) || input.vecY < 0 || input.radius <= 0
				|| input.image.length == 0) {
			main.showError("Please fill all fields.");
			return;
		} else {
			main.hideError();
		}

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
	
	serverMessage: function(e) {
		var data = e.data;
		console.log(data);
		if (data.substr(0, 1) == "{") {
			data = JSON.parse(data);

			// response
			main.divResult.html('<img src="/img/'+ data.Image +'" alt="" />')
		} else {
			process.divLoader.width(parseInt(data) + "%");
		}
	},

	stop: function() {
		if (process.connection == null) {
			return;
		}
		process.connection.send("stop");
	}
	
};