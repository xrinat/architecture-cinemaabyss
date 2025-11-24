using EventsService.Kafka;
using EventsService.Models;
using Microsoft.AspNetCore.Mvc;

namespace EventsService.Controllers
{
    [ApiController]
    [Route("api/events")]
    public class EventsController : ControllerBase
    {
        private readonly EventProducer _producer;
        private readonly ILogger<EventsController> _logger;

        public EventsController(EventProducer producer, ILogger<EventsController> logger)
        {
            _producer = producer;
            _logger = logger;
        }
        
        // GET /api/events/health
        [HttpGet("health")]
        [ProducesResponseType(typeof(object), 200)]
        public IActionResult GetHealth()
        {
            _logger.LogInformation("Health check requested.");
            return Ok(new { status = true });
        }

        // POST /api/events/movie
        [HttpPost("movie")]
        [ProducesResponseType(typeof(EventResponse), 201)]
        [ProducesResponseType(500)]
        public async Task<IActionResult> CreateMovieEvent([FromBody] MovieEvent movieEvent)
        {
            try
            {
                var result = await _producer.ProduceAsync("movie-events", movieEvent);
                return StatusCode(201, result);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to produce movie event to Kafka.");
                return StatusCode(500, new { error = "Internal Server Error during movie event production." });
            }
        }

        // POST /api/events/user
        [HttpPost("user")]
        [ProducesResponseType(typeof(EventResponse), 201)]
        [ProducesResponseType(500)]
        public async Task<IActionResult> CreateUserEvent([FromBody] UserEvent userEvent)
        {
            try
            {
                var result = await _producer.ProduceAsync("user-events", userEvent);
                return StatusCode(201, result);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to produce user event to Kafka.");
                return StatusCode(500, new { error = "Internal Server Error during user event production." });
            }
        }

        // POST /api/events/payment
        [HttpPost("payment")]
        [ProducesResponseType(typeof(EventResponse), 201)]
        [ProducesResponseType(500)]
        public async Task<IActionResult> CreatePaymentEvent([FromBody] PaymentEvent paymentEvent)
        {
            try
            {
                var result = await _producer.ProduceAsync("payment-events", paymentEvent);
                return StatusCode(201, result);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to produce payment event to Kafka.");
                return StatusCode(500, new { error = "Internal Server Error during payment event production." });
            }
        }
    }
}