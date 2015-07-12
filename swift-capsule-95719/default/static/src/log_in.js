function LogInCtrl($scope, $http) {
	
	$scope.log_in = function() {
		
		var new_auth = {
			username: $scope.username,
			password: $scope.password,
		};
		
		//console.log(new_auth);
		
		$http.post('/log_in', new_auth).
		success(function(data, status, headers, config) {
			
			if (typeof(sessionStorage) === "undefined") {
				$scope.log_in_status = "Error: Your browser doesn't support session storage.";
				return;
			}
			
			sessionStorage.setItem("auth_jwt", data);
			
			$scope.log_in_status = "Log in successful.";
			
		}).
		error(function(data, status, headers, config) {
			
			$scope.log_in_status = "Log in failed: " + data;
		});
	};
}

app.controller('LogInCtrl', LogInCtrl);
