function WebsocketClient() {
	this.ws;
}

WebsocketClient.prototype.Connect = function(url, $scope) {
	
	$scope.connection_status = "Not connected";
	
	var ws = this.ws = new WebSocket(url);
	
	var that = this;

	window.onbeforeunload = function() {
		that.ws.close();
	};
	
	this.ws.onopen = function(e) {
		$scope.connection_status = "Connected";
		$scope.$apply();
		console.log("Connected to websocket server: " + url);
		that.ws.send(JSON.stringify({msg: "hi there!"}));
	};
	
	this.ws.onmessage = function(e) {
		
		if (e.data === "{}") {
			that.ws.onmessage = function() {};
			that.ws.close();
			return;
		}
		
		console.log("Websocket msg: " + e.data);
		var data = JSON.parse(e.data);
		$scope.pidgeon_capacity = "Pidgeon Capacity: " + data.msg + "%";
		$scope.$apply();
	}
	
	this.ws.onerror = function(e) {
		//console.log("Websocket error: " + JSON.stringify(e, null, 4));
		that.ws.close();
	}
	
	this.ws.onclose = function(e) {
		
		that.ws.onclose = function() {};
		
		$scope.connection_status = "Not connected";
		$scope.$apply();
		console.log("Websocket connection closed. Attempting to reconnect in 10 seconds...");
		window.setTimeout(function() { that.Connect(url, $scope); }, 10000);
	}
};
