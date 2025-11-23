using System.Text.Json.Serialization;

namespace EventsService.Models
{
    // Базовый класс для всех событий
    public abstract class BaseEvent
    {
        public Guid EventId { get; set; } = Guid.NewGuid();
        public DateTime Timestamp { get; set; } = DateTime.UtcNow;
        public abstract string EventType { get; }
    }

    // Событие фильма (соответствует /api/events/movie)
    public class MovieEvent : BaseEvent
    {
        [JsonIgnore]
        public override string EventType => "Movie";
        
        public int MovieId { get; set; }
        public string Title { get; set; } = string.Empty;
        public string Action { get; set; } = string.Empty; // e.g., "viewed", "updated"
        public int? UserId { get; set; } // Кто совершил действие
    }

    // Событие пользователя (соответствует /api/events/user)
    public class UserEvent : BaseEvent
    {
        [JsonIgnore]
        public override string EventType => "User";

        public int UserId { get; set; }
        public string Action { get; set; } = string.Empty; // e.g., "registered", "logged_in"
        public string? Username { get; set; }
        public string? Email { get; set; }
    }

    // Событие платежа (соответствует /api/events/payment)
    public class PaymentEvent : BaseEvent
    {
        [JsonIgnore]
        public override string EventType => "Payment";

        public int PaymentId { get; set; }
        public int UserId { get; set; }
        public decimal Amount { get; set; }
        public string Status { get; set; } = string.Empty; // e.g., "completed", "failed"
    }

    // Ответ API
    public class EventResponse
    {
        public string Status { get; set; } = "success";
        public int Partition { get; set; }
        public long Offset { get; set; }
        public BaseEvent Event { get; set; }
    }
}