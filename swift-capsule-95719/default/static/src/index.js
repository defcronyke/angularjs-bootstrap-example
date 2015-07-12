var app = angular.module('app', ['ui.bootstrap', 'ngRoute']);

app.config(['$routeProvider', '$locationProvider', function($routeProvider, $locationProvider) {
	
	$routeProvider.
		when("/", { templateUrl: '/static/templates/main_partial.html', controller: MainCtrl }).
		when("/contact", { templateUrl: '/static/templates/contact_partial.html', controller: ContactCtrl }).
		when("/log_in", { templateUrl: '/static/templates/log_in_partial.html', controller: LogInCtrl }).
		when("/register", { templateUrl: '/static/templates/register_partial.html', controller: RegisterCtrl }).
		otherwise({ redirectTo: "/" });
	
	$locationProvider.html5Mode(false);
	
}]);

function BodyCtrl($scope, $location) {
	
	$scope.ws = new WebsocketClient();
    //$scope.ws.Connect("ws://130.211.161.241:8195", $scope);		// Development Websocket server
    $scope.ws.Connect("ws://107.178.215.195:8195", $scope);		// Production Websocket server
}

app.controller('BodyCtrl', BodyCtrl);

function HeaderCtrl($scope, $location) {
	
	$scope.isActive = function(viewLocation) { 
        return viewLocation === $location.path();
    };
}

app.controller('HeaderCtrl', HeaderCtrl);

function MainCtrl($scope) {
	
}

function ContactCtrl($scope) {
	
}

function LogInCtrl($scope) {
	
}

function RegisterCtrl($scope) {
	
}

function CarouselCtrl($scope) {
	$scope.myInterval = 5000;
	var slides = $scope.slides = [];
	$scope.addSlide = function() {
		var newWidth = 600 + slides.length + 1;
		slides.push({
			image: 'http://lorempixel.com/' + newWidth + '/300',
			text: ['More','Extra','Lots of','Surplus'][slides.length % 4] + ' ' +
				  ['Person', 'Situation', 'Image', 'Penguins'][slides.length % 4]
		});
	};
	
	for (var i=0; i<4; i++) {
		$scope.addSlide();
	}
}

app.controller('CarouselCtrl', CarouselCtrl);
