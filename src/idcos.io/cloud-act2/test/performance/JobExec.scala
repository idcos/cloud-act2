package act2

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._


class JobExec extends Simulation {

    val httpProtocol = http.baseUrl("http://192.168.1.17:6868")

    var stdUser = scenario("Standard User")
        .exec(http("Access act2 master").get("/"))
        .pause(1, 2)

    setUp(
        stdUser.inject(atOnceUsers(10000).protocols(httpProtocol),
    )

}
